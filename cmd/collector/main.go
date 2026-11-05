package main

import (
	"log/slog"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/Richonn/wf-market-tracker/internal/alert"
	"github.com/Richonn/wf-market-tracker/internal/api"
	"github.com/Richonn/wf-market-tracker/internal/collector"
	"github.com/Richonn/wf-market-tracker/internal/db"
	"github.com/Richonn/wf-market-tracker/internal/model"
)

const collectInterval = 45 * time.Minute

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	dsn := mustEnv("DATABASE_URL")

	database, err := db.New(dsn)
	if err != nil {
		slog.Error("db connect", "err", err)
		os.Exit(1)
	}
	defer database.Close()

	alertEngine := loadAlertEngine(database)

	client := api.NewClient()
	c := collector.New(client, database, alertEngine)

	slog.Info("bootstrapping historical data...")
	if err := c.Bootstrap(); err != nil {
		slog.Error("bootstrap failed", "err", err)
		os.Exit(1)
	}

	slog.Info("starting collection loop", "interval", collectInterval)
	c.Run(collectInterval)
}

func loadAlertEngine(database *db.DB) *alert.Engine {
	path := os.Getenv("ALERTS_CONFIG")
	if path == "" {
		path = "alerts.yaml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		slog.Warn("no alerts config found, alerts disabled", "path", path)
		return nil
	}

	var cfg model.AlertConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		slog.Error("parse alerts config", "err", err)
		return nil
	}

	var senders []alert.Sender
	if v := os.Getenv("NTFY_URL"); v != "" {
		senders = append(senders, &alert.Ntfy{URL: v, Token: os.Getenv("NTFY_TOKEN")})
	}
	if v := os.Getenv("DISCORD_WEBHOOK_URL"); v != "" {
		senders = append(senders, &alert.Discord{WebhookURL: v})
	}

	if len(senders) == 0 {
		slog.Warn("alerts config loaded but no sender configured (set NTFY_URL or DISCORD_WEBHOOK_URL)")
	}

	return alert.NewEngine(database, &cfg, senders...)
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("missing required env var", "key", key)
		os.Exit(1)
	}
	return v
}
