package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Richonn/wf-market-tracker/internal/db"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	database, err := db.NewReadOnly(mustEnv("DATABASE_URL"))
	if err != nil {
		slog.Error("db connect", "err", err)
		os.Exit(1)
	}
	defer database.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /items", handleItems(database))
	mux.HandleFunc("GET /items/{slug}/candles", handleCandles(database))
	mux.HandleFunc("GET /movers", handleMovers(database))
	mux.Handle("/", http.FileServer(http.Dir("/web")))

	handler := chain(mux, securityHeaders, rateLimiter, accessLog)

	addr := ":8080"
	slog.Info("API listening", "addr", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func handleItems(d *db.DB) http.HandlerFunc {
	return cached(5*time.Minute, func(w http.ResponseWriter, r *http.Request) {
		items, err := d.ListItems()
		if err != nil {
			writeError(w, http.StatusInternalServerError)
			slog.Error("list items", "err", err)
			return
		}
		writeJSON(w, items)
	})
}

func handleCandles(d *db.DB) http.HandlerFunc {
	return cached(5*time.Minute, func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		interval := r.URL.Query().Get("interval")
		if interval == "" {
			interval = "1h"
		}
		candles, err := d.ItemCandles(slug, interval)
		if err != nil {
			writeError(w, http.StatusInternalServerError)
			slog.Error("candles", "slug", slug, "err", err)
			return
		}
		writeJSON(w, candles)
	})
}

func handleMovers(d *db.DB) http.HandlerFunc {
	return cached(5*time.Minute, func(w http.ResponseWriter, r *http.Request) {
		movers, err := d.Movers()
		if err != nil {
			writeError(w, http.StatusInternalServerError)
			slog.Error("movers", "err", err)
			return
		}
		writeJSON(w, movers)
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encode response", "err", err)
	}
}

func writeError(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'unsafe-inline' https://unpkg.com; style-src 'unsafe-inline'; connect-src 'self'; img-src 'self'")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}

func accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote", r.RemoteAddr,
		)
	})
}

type rateLimiterMiddleware struct {
	mu      sync.Mutex
	clients map[string][]time.Time
}

var limiter = &rateLimiterMiddleware{clients: make(map[string][]time.Time)}

func rateLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := strings.Split(r.RemoteAddr, ":")[0]
		now := time.Now()
		window := now.Add(-time.Minute)

		limiter.mu.Lock()
		times := limiter.clients[ip]
		var recent []time.Time
		for _, t := range times {
			if t.After(window) {
				recent = append(recent, t)
			}
		}
		recent = append(recent, now)
		limiter.clients[ip] = recent
		count := len(recent)
		limiter.mu.Unlock()

		if count > 60 {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type cacheEntry struct {
	body      []byte
	expiresAt time.Time
}

func cached(ttl time.Duration, h http.HandlerFunc) http.HandlerFunc {
	var mu sync.Mutex
	var entry *cacheEntry
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.String()
		mu.Lock()
		if entry != nil && entry.expiresAt.After(time.Now()) && r.URL.String() == key {
			body := entry.body
			mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.Write(body)
			return
		}
		mu.Unlock()

		rec := &responseRecorder{header: make(http.Header)}
		h(rec, r)

		mu.Lock()
		entry = &cacheEntry{body: rec.body, expiresAt: time.Now().Add(ttl)}
		mu.Unlock()

		for k, v := range rec.header {
			w.Header()[k] = v
		}
		w.WriteHeader(rec.code)
		w.Write(rec.body)
	}
}

type responseRecorder struct {
	header http.Header
	body   []byte
	code   int
}

func (r *responseRecorder) Header() http.Header        { return r.header }
func (r *responseRecorder) WriteHeader(code int)       { r.code = code }
func (r *responseRecorder) Write(b []byte) (int, error) { r.body = append(r.body, b...); return len(b), nil }

func chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("missing required env var", "key", key)
		os.Exit(1)
	}
	return v
}
