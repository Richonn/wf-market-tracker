package collector

import (
	"log/slog"
	"sort"
	"time"

	"github.com/Richonn/wf-market-tracker/internal/api"
	"github.com/Richonn/wf-market-tracker/internal/db"
	"github.com/Richonn/wf-market-tracker/internal/model"
)

type Collector struct {
	api         *api.Client
	db          *db.DB
	lastRetention time.Time
}

func New(api *api.Client, db *db.DB) *Collector {
	return &Collector{api: api, db: db}
}

func (c *Collector) Bootstrap() error {
	items, err := c.api.GetItems()
	if err != nil {
		return err
	}

	for _, item := range items {
		if !item.BulkTradable {
			continue
		}
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
		if !item.BulkTradable {
			continue
		}

		orders, err := c.api.GetTopOrders(item.Slug)
		if err != nil {
			slog.Warn("get orders", "slug", item.Slug, "err", err)
			continue
		}

		snapshot := ComputeSnapshot(item.Slug, orders)
		if err := c.db.InsertSnapshot(snapshot); err != nil {
			slog.Error("insert snapshot", "slug", item.Slug, "err", err)
		}
	}
	return nil
}

func ComputeSnapshot(slug string, orders []model.Order) model.Snapshot {
	var sells, buys []int

	for _, o := range orders {
		if o.UserStatus == "offline" {
			continue
		}
		switch o.OrderType {
		case "sell":
			sells = append(sells, o.Platinum)
		case "buy":
			buys = append(buys, o.Platinum)
		}
	}

	sort.Ints(sells)
	sort.Sort(sort.Reverse(sort.IntSlice(buys)))

	s := model.Snapshot{
		ItemSlug:         slug,
		CollectedAt:      time.Now().UTC(),
		ActiveSellOrders: len(sells),
		ActiveBuyOrders:  len(buys),
	}
	if len(sells) > 0 {
		s.SellMin = sells[0]
		s.SellMedian = sells[len(sells)/2]
	}
	if len(buys) > 0 {
		s.BuyMax = buys[0]
	}
	return s
}
