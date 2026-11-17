package collector

import (
	"log/slog"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/Richonn/wf-market-tracker/internal/alert"
	"github.com/Richonn/wf-market-tracker/internal/api"
	"github.com/Richonn/wf-market-tracker/internal/db"
	"github.com/Richonn/wf-market-tracker/internal/model"
)

type Collector struct {
	api           *api.Client
	db            *db.DB
	alerts        *alert.Engine
	lastRetention time.Time
}

func New(api *api.Client, db *db.DB, alerts *alert.Engine) *Collector {
	return &Collector{api: api, db: db, alerts: alerts}
}

func (c *Collector) Bootstrap() error {
	items, err := c.api.GetItems()
	if err != nil {
		return err
	}

	for _, item := range items {
		if err := c.db.UpsertItem(item); err != nil {
			slog.Error("upsert item", "slug", item.Slug, "err", err)
			continue
		}

		seeded, err := c.db.ItemSeeded(item.Slug)
		if err != nil || seeded {
			continue
		}

		stats, err := c.api.GetStatistics(item.Slug)
		if err != nil {
			slog.Warn("seed stats", "slug", item.Slug, "err", err)
			continue
		}

		for _, s := range stats {
			if s.OrderType != "sell" {
				continue
			}
			if err := c.db.InsertHistoricalSnapshot(item.Slug, s.Datetime, s.Median, s.MinPrice, s.MaxPrice, s.Volume); err != nil {
				slog.Error("insert historical", "slug", item.Slug, "err", err)
			}
		}
	}
	return nil
}

func (c *Collector) Run(interval time.Duration) {
	for {
		start := time.Now()
		slog.Info("collection cycle started")

		if err := c.collect(); err != nil {
			slog.Error("collection cycle failed", "err", err)
		} else {
			slog.Info("collection cycle done", "duration", time.Since(start).Round(time.Second))
			c.pingHealthcheck()
		}

		if c.alerts != nil {
			c.alerts.Evaluate()
		}

		if time.Since(c.lastRetention) >= 24*time.Hour {
			if err := c.db.RunRetention(); err != nil {
				slog.Error("retention failed", "err", err)
			} else {
				c.lastRetention = time.Now()
				slog.Info("retention done")
			}
		}

		time.Sleep(interval - time.Since(start))
	}
}

func (c *Collector) collect() error {
	items, err := c.api.GetItems()
	if err != nil {
		return err
	}

	for _, item := range items {
		orders, err := c.api.GetTopOrders(item.Slug)
		if err != nil {
			slog.Warn("get orders", "slug", item.Slug, "err", err)
			continue
		}

		for _, snapshot := range ComputeSnapshots(item.Slug, orders) {
			if err := c.db.InsertSnapshot(snapshot); err != nil {
				slog.Error("insert snapshot", "slug", item.Slug, "rank", snapshot.Rank, "err", err)
			}
		}
	}
	return nil
}

func (c *Collector) pingHealthcheck() {
	url := os.Getenv("HEALTHCHECKS_URL")
	if url == "" {
		return
	}
	resp, err := http.Get(url)
	if err != nil {
		slog.Warn("healthcheck ping failed", "err", err)
		return
	}
	resp.Body.Close()
}

func ComputeSnapshots(slug string, orders []model.Order) []model.Snapshot {
	type rankOrders struct {
		sells []int
		buys  []int
	}
	byRank := map[int]*rankOrders{}

	for _, o := range orders {
		if o.User.Status == "offline" {
			continue
		}
		if _, ok := byRank[o.Rank]; !ok {
			byRank[o.Rank] = &rankOrders{}
		}
		switch o.Type {
		case "sell":
			byRank[o.Rank].sells = append(byRank[o.Rank].sells, o.Platinum)
		case "buy":
			byRank[o.Rank].buys = append(byRank[o.Rank].buys, o.Platinum)
		}
	}

	now := time.Now().UTC()
	var snapshots []model.Snapshot
	for rank, ro := range byRank {
		sort.Ints(ro.sells)
		sort.Sort(sort.Reverse(sort.IntSlice(ro.buys)))

		s := model.Snapshot{
			ItemSlug:         slug,
			CollectedAt:      now,
			Rank:             rank,
			ActiveSellOrders: len(ro.sells),
			ActiveBuyOrders:  len(ro.buys),
		}
		if len(ro.sells) > 0 {
			s.SellMin = ro.sells[0]
			s.SellMedian = ro.sells[len(ro.sells)/2]
		}
		if len(ro.buys) > 0 {
			s.BuyMax = ro.buys[0]
		}
		snapshots = append(snapshots, s)
	}
	return snapshots
}

func ComputeSnapshot(slug string, orders []model.Order) model.Snapshot {
	snaps := ComputeSnapshots(slug, orders)
	for _, s := range snaps {
		if s.Rank == 0 {
			return s
		}
	}
	return model.Snapshot{ItemSlug: slug, CollectedAt: time.Now().UTC()}
}
