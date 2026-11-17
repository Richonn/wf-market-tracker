package db

import "time"

func (d *DB) RunRetention() error {
	if err := d.aggregateHourly(); err != nil {
		return err
	}
	if err := d.aggregateDaily(); err != nil {
		return err
	}
	if err := d.pruneRaw(); err != nil {
		return err
	}
	return nil
}

func (d *DB) aggregateHourly() error {
	_, err := d.Exec(`
		INSERT INTO price_snapshots_hourly
			(item_slug, hour, rank, sell_min, sell_median, buy_max, avg_sell_orders, avg_buy_orders, sample_count)
		SELECT
			item_slug,
			date_trunc('hour', collected_at) AS hour,
			rank,
			MIN(sell_min),
			PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY sell_median)::INT,
			MAX(buy_max),
			AVG(active_sell_orders)::INT,
			AVG(active_buy_orders)::INT,
			COUNT(*)
		FROM price_snapshots
		WHERE collected_at < date_trunc('hour', NOW())
		  AND collected_at >= NOW() - INTERVAL '32 days'
		GROUP BY item_slug, date_trunc('hour', collected_at), rank
		ON CONFLICT (item_slug, hour, rank) DO NOTHING
	`)
	return err
}

func (d *DB) aggregateDaily() error {
	_, err := d.Exec(`
		INSERT INTO price_snapshots_daily
			(item_slug, day, rank, sell_min, sell_median, buy_max, avg_sell_orders, avg_buy_orders, sample_count)
		SELECT
			item_slug,
			date_trunc('day', hour)::DATE AS day,
			rank,
			MIN(sell_min),
			PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY sell_median)::INT,
			MAX(buy_max),
			AVG(avg_sell_orders)::INT,
			AVG(avg_buy_orders)::INT,
			SUM(sample_count)
		FROM price_snapshots_hourly
		WHERE hour < date_trunc('day', NOW())
		GROUP BY item_slug, date_trunc('day', hour)::DATE, rank
		ON CONFLICT (item_slug, day, rank) DO NOTHING
	`)
	return err
}

func (d *DB) pruneRaw() error {
	cutoff := time.Now().UTC().Add(-30 * 24 * time.Hour)
	_, err := d.Exec(`DELETE FROM price_snapshots WHERE collected_at < $1`, cutoff)
	return err
}
