package api

import (
	"encoding/json"
	"fmt"
	"time"
)

type StatEntry struct {
	Datetime   time.Time `json:"datetime"`
	AvgPrice   float64   `json:"avg_price"`
	MinPrice   float64   `json:"min_price"`
	MaxPrice   float64   `json:"max_price"`
	Median     float64   `json:"median"`
	Volume     int       `json:"volume"`
	OrderType  string    `json:"order_type"`
}

type statisticsResponse struct {
	Payload struct {
		StatsClosed struct {
			Days90 []StatEntry `json:"90days"`
		} `json:"statistics_closed"`
	} `json:"payload"`
}

func (c *Client) GetStatistics(slug string) ([]StatEntry, error) {
	body, err := c.get(fmt.Sprintf("%s/items/%s/statistics", baseV1, slug))
	if err != nil {
		return nil, err
	}

	var resp statisticsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return resp.Payload.StatsClosed.Days90, nil
}
