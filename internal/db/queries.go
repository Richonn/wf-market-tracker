package db

import (
	"time"

	"github.com/Richonn/wf-market-tracker/internal/model"
)

func (d *DB) UpsertItem(item model.Item) error {
	_, err := d.Exec(`
		INSERT INTO items (id, slug, name, tags)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, tags = EXCLUDED.tags
	`, item.ID, item.Slug, item.Name, item.Tags)
	return err
}

func (d *DB) InsertSnapshot(s model.Snapshot) error {
	_, err := d.Exec(`
		INSERT INTO price_snapshots
			(item_slug, collected_at, sell_min, sell_median, buy_max, active_sell_orders, active_buy_orders)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, s.ItemSlug, s.CollectedAt, s.SellMin, s.SellMedian, s.BuyMax, s.ActiveSellOrders, s.ActiveBuyOrders)
	return err
}

func (d *DB) ItemSeeded(slug string) (bool, error) {
	var count int
	err := d.QueryRow(`SELECT COUNT(*) FROM price_snapshots WHERE item_slug = $1`, slug).Scan(&count)
	return count > 0, err
}

func (d *DB) InsertHistoricalSnapshot(slug string, t time.Time, median, min, max float64, volume int) error {
	_, err := d.Exec(`
		INSERT INTO price_snapshots
			(item_slug, collected_at, sell_min, sell_median, buy_max, active_sell_orders, active_buy_orders)
		VALUES ($1, $2, $3, $4, $5, $6, 0)
		ON CONFLICT DO NOTHING
	`, slug, t, int(min), int(median), int(max), volume)
	return err
}
