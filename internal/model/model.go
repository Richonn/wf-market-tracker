package model

import "time"

type ItemI18n struct {
	En struct {
		Name string `json:"name"`
	} `json:"en"`
}

type Item struct {
	ID           string    `json:"id"`
	Slug         string    `json:"slug"`
	I18n         ItemI18n  `json:"i18n"`
	BulkTradable bool      `json:"bulkTradable"`
	Tags         []string  `json:"tags"`
}

func (i Item) Name() string {
	return i.I18n.En.Name
}

type OrderUser struct {
	Status string `json:"status"`
}

type Order struct {
	ID       string    `json:"id"`
	Type     string    `json:"type"`
	Platinum int       `json:"platinum"`
	Quantity int       `json:"quantity"`
	Rank     int       `json:"rank"`
	User     OrderUser `json:"user"`
}

type Snapshot struct {
	ID               int64     `db:"id"`
	ItemSlug         string    `db:"item_slug"`
	CollectedAt      time.Time `db:"collected_at"`
	Rank             int       `db:"rank"`
	SellMin          int       `db:"sell_min"`
	SellMedian       int       `db:"sell_median"`
	BuyMax           int       `db:"buy_max"`
	ActiveSellOrders int       `db:"active_sell_orders"`
	ActiveBuyOrders  int       `db:"active_buy_orders"`
}
