package db

import "time"

type ItemRow struct {
	Slug        string  `json:"slug"`
	Name        string  `json:"name"`
	SellMedian  int     `json:"sell_median"`
	Change24hPct float64 `json:"change_24h_pct"`
	Change7dPct  float64 `json:"change_7d_pct"`
}

type Candle struct {
	Time       time.Time `json:"time"`
	SellMin    int       `json:"sell_min"`
	SellMedian int       `json:"sell_median"`
	BuyMax     int       `json:"buy_max"`
	Volume     int       `json:"volume"`
}

func (d *DB) ListItems() ([]ItemRow, error) {
	rows, err := d.Query(`
		SELECT
			i.slug,
			i.name,
			latest.sell_median,
			ROUND(
				(latest.sell_median - prev24h.sell_median)::numeric
				/ NULLIF(prev24h.sell_median, 0) * 100, 1
			) AS change_24h_pct,
			ROUND(
				(latest.sell_median - prev7d.sell_median)::numeric
				/ NULLIF(prev7d.sell_median, 0) * 100, 1
			) AS change_7d_pct
		FROM items i
		JOIN LATERAL (
			SELECT sell_median FROM price_snapshots
			WHERE item_slug = i.slug ORDER BY collected_at DESC LIMIT 1
		) latest ON true
		LEFT JOIN LATERAL (
			SELECT sell_median FROM price_snapshots
			WHERE item_slug = i.slug AND collected_at <= NOW() - INTERVAL '24 hours'
			ORDER BY collected_at DESC LIMIT 1
		) prev24h ON true
		LEFT JOIN LATERAL (
			SELECT sell_median FROM price_snapshots
			WHERE item_slug = i.slug AND collected_at <= NOW() - INTERVAL '7 days'
			ORDER BY collected_at DESC LIMIT 1
		) prev7d ON true
		WHERE latest.sell_median > 0
		ORDER BY i.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ItemRow
	for rows.Next() {
		var item ItemRow
		if err := rows.Scan(&item.Slug, &item.Name, &item.SellMedian, &item.Change24hPct, &item.Change7dPct); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (d *DB) ItemCandles(slug, interval string) ([]Candle, error) {
	var query string
	switch interval {
	case "1d":
		query = `
			SELECT day::timestamptz, sell_min, sell_median, buy_max, sample_count
			FROM price_snapshots_daily
			WHERE item_slug = $1
			ORDER BY day DESC LIMIT 365`
	default:
		query = `
			SELECT hour, sell_min, sell_median, buy_max, sample_count
			FROM price_snapshots_hourly
			WHERE item_slug = $1
			ORDER BY hour DESC LIMIT 720`
	}

	rows, err := d.Query(query, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candles []Candle
	for rows.Next() {
		var c Candle
		if err := rows.Scan(&c.Time, &c.SellMin, &c.SellMedian, &c.BuyMax, &c.Volume); err != nil {
			return nil, err
		}
		candles = append(candles, c)
	}
	return candles, rows.Err()
}

func (d *DB) Movers() ([]ItemRow, error) {
	rows, err := d.Query(`
		SELECT
			i.slug,
			i.name,
			latest.sell_median,
			ROUND(
				(latest.sell_median - prev24h.sell_median)::numeric
				/ NULLIF(prev24h.sell_median, 0) * 100, 1
			) AS change_24h_pct,
			0::numeric AS change_7d_pct
		FROM items i
		JOIN LATERAL (
			SELECT sell_median FROM price_snapshots
			WHERE item_slug = i.slug ORDER BY collected_at DESC LIMIT 1
		) latest ON true
		JOIN LATERAL (
			SELECT sell_median FROM price_snapshots
			WHERE item_slug = i.slug AND collected_at <= NOW() - INTERVAL '24 hours'
			ORDER BY collected_at DESC LIMIT 1
		) prev24h ON true
		WHERE latest.sell_median > 0 AND prev24h.sell_median > 0
		ORDER BY ABS(
			(latest.sell_median - prev24h.sell_median)::numeric
			/ NULLIF(prev24h.sell_median, 0)
		) DESC
		LIMIT 20
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ItemRow
	for rows.Next() {
		var item ItemRow
		if err := rows.Scan(&item.Slug, &item.Name, &item.SellMedian, &item.Change24hPct, &item.Change7dPct); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
