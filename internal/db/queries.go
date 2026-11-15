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
	`, item.ID, item.Slug, item.Name(), item.Tags)
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

func (d *DB) AllSlugs() ([]string, error) {
	rows, err := d.Query(`SELECT slug FROM items ORDER BY slug`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var slugs []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		slugs = append(slugs, s)
	}
	return slugs, rows.Err()
}

func (d *DB) MovingAvgSellMedian(slug string, windowDays int) (float64, error) {
	var avg float64
	err := d.QueryRow(`
		SELECT COALESCE(AVG(sell_median), 0)
		FROM price_snapshots
		WHERE item_slug = $1
		  AND sell_median > 0
		  AND collected_at >= NOW() - ($2 || ' days')::INTERVAL
	`, slug, windowDays).Scan(&avg)
	return avg, err
}

func (d *DB) LatestSnapshot(slug string) (model.Snapshot, error) {
	var s model.Snapshot
	err := d.QueryRow(`
		SELECT item_slug, collected_at, sell_min, sell_median, buy_max, active_sell_orders, active_buy_orders
		FROM price_snapshots WHERE item_slug = $1
		ORDER BY collected_at DESC LIMIT 1
	`, slug).Scan(&s.ItemSlug, &s.CollectedAt, &s.SellMin, &s.SellMedian, &s.BuyMax, &s.ActiveSellOrders, &s.ActiveBuyOrders)
	return s, err
}

func (d *DB) SnapshotAt(slug string, t time.Time) (model.Snapshot, error) {
	var s model.Snapshot
	err := d.QueryRow(`
		SELECT item_slug, collected_at, sell_min, sell_median, buy_max, active_sell_orders, active_buy_orders
		FROM price_snapshots WHERE item_slug = $1 AND collected_at <= $2
		ORDER BY collected_at DESC LIMIT 1
	`, slug, t).Scan(&s.ItemSlug, &s.CollectedAt, &s.SellMin, &s.SellMedian, &s.BuyMax, &s.ActiveSellOrders, &s.ActiveBuyOrders)
	return s, err
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
