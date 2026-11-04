package model

import "time"

type Item struct {
	ID           string   `json:"id"`
	Slug         string   `json:"slug"`
	Name         string   `json:"name"`
	BulkTradable bool     `json:"bulkTradable"`
	Tags         []string `json:"tags"`
}

type Order struct {
	ID         string `json:"id"`
	ItemSlug   string `json:"itemSlug"`
	OrderType  string `json:"orderType"`
	Platinum   int    `json:"platinum"`
	Quantity   int    `json:"quantity"`
	UserName   string `json:"userName"`
	UserStatus string `json:"userStatus"`
}

type Snapshot struct {
	ID               int64     `db:"id"`
	ItemSlug         string    `db:"item_slug"`
	CollectedAt      time.Time `db:"collected_at"`
	SellMin          int       `db:"sell_min"`
	SellMedian       int       `db:"sell_median"`
	BuyMax           int       `db:"buy_max"`
	ActiveSellOrders int       `db:"active_sell_orders"`
	ActiveBuyOrders  int       `db:"active_buy_orders"`
}
