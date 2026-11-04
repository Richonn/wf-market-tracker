package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/Richonn/wf-market-tracker/internal/api"
	"github.com/Richonn/wf-market-tracker/internal/collector"
	"github.com/Richonn/wf-market-tracker/internal/db"
)

const collectInterval = 45 * time.Minute

func main() {
	dsn := mustEnv("DATABASE_URL")

	database, err := db.New(dsn)
	if err != nil {
		slog.Error("db connect", "err", err)
		os.Exit(1)
	}
	defer database.Close()

	client := api.NewClient()
	c := collector.New(client, database)

	slog.Info("bootstrapping historical data...")
	if err := c.Bootstrap(); err != nil {
		slog.Error("bootstrap failed", "err", err)
		os.Exit(1)
	}

	slog.Info("starting collection loop", "interval", collectInterval)
	c.Run(collectInterval)
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("missing required env var", "key", key)
		os.Exit(1)
	}
	return v
}
