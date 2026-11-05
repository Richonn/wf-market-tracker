package alert

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Richonn/wf-market-tracker/internal/db"
	"github.com/Richonn/wf-market-tracker/internal/model"
)

type Sender interface {
	Send(title, message string) error
}

type Engine struct {
	db       *db.DB
	senders  []Sender
	config   *model.AlertConfig
	state    *model.AlertState
	cooldown time.Duration
}

func NewEngine(d *db.DB, cfg *model.AlertConfig, senders ...Sender) *Engine {
	return &Engine{
		db:       d,
		senders:  senders,
		config:   cfg,
		state:    model.NewAlertState(),
		cooldown: time.Duration(cfg.CooldownMinutes) * time.Minute,
	}
}

func (e *Engine) Evaluate() {
	slugs, err := e.db.AllSlugs()
	if err != nil {
		slog.Error("alert: fetch slugs", "err", err)
		return
	}
	fired := 0
	max := e.config.MaxPerCycle
	if max <= 0 {
		max = 10
	}
	for _, slug := range slugs {
		if fired >= max {
			slog.Warn("alert: max per cycle reached, skipping remaining", "max", max)
			break
		}
		fired += e.evaluateSlug(slug)
	}
}

func (e *Engine) evaluateSlug(slug string) int {
	fired := 0

	snapshot, err := e.db.LatestSnapshot(slug)
	if err != nil || snapshot.SellMedian == 0 {
		return 0
	}

	if e.config.SellBelowAvgPct != 0 {
		avg, err := e.db.MovingAvgSellMedian(slug, e.config.TrendWindowDays)
		if err == nil && avg > 0 {
			threshold := avg * (1 + e.config.SellBelowAvgPct/100)
			if float64(snapshot.SellMedian) < threshold {
				key := fmt.Sprintf("%s:sell_below_avg", slug)
				if e.state.CanFire(key, e.cooldown) {
					e.fire(
						fmt.Sprintf("Price alert — %s", slug),
						fmt.Sprintf(
							"Sell median %dpt is %.1f%% below %dd average (avg %.0fpt, threshold %.0fpt)",
							snapshot.SellMedian, e.config.SellBelowAvgPct, e.config.TrendWindowDays, avg, threshold,
						),
					)
					e.state.Record(key)
					fired++
				}
			}
		}
	}

	if e.config.Change48hPct != 0 {
		prev, err := e.db.SnapshotAt(slug, time.Now().Add(-48*time.Hour))
		if err == nil && prev.SellMedian > 0 {
			change := float64(snapshot.SellMedian-prev.SellMedian) / float64(prev.SellMedian) * 100
			if (e.config.Change48hPct < 0 && change <= e.config.Change48hPct) ||
				(e.config.Change48hPct > 0 && change >= e.config.Change48hPct) {
				key := fmt.Sprintf("%s:change_48h", slug)
				if e.state.CanFire(key, e.cooldown) {
					e.fire(
						fmt.Sprintf("Variation alert — %s", slug),
						fmt.Sprintf("%.1f%% in 48h (threshold %.1f%%)", change, e.config.Change48hPct),
					)
					e.state.Record(key)
					fired++
				}
			}
		}
	}

	return fired
}

func (e *Engine) fire(title, message string) {
	for _, s := range e.senders {
		if err := s.Send(title, message); err != nil {
			slog.Error("alert send", "err", err)
		}
	}
}
