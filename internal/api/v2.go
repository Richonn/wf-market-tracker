package api

import (
	"encoding/json"
	"fmt"

	"github.com/Richonn/wf-market-tracker/internal/model"
)

type itemsResponse struct {
	Data []model.Item `json:"data"`
}

type ordersResponse struct {
	Data json.RawMessage `json:"data"`
}

type ordersData struct {
	Buy  []model.Order `json:"buy"`
	Sell []model.Order `json:"sell"`
}

func (c *Client) GetItems() ([]model.Item, error) {
	body, err := c.get(fmt.Sprintf("%s/items", baseV2))
	if err != nil {
		return nil, err
	}

	var resp itemsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) GetTopOrders(slug string) ([]model.Order, error) {
	body, err := c.get(fmt.Sprintf("%s/orders/item/%s/top", baseV2, slug))
	if err != nil {
		return nil, err
	}

	var resp ordersResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	// Try array format first
	var orders []model.Order
	if err := json.Unmarshal(resp.Data, &orders); err == nil {
		return orders, nil
	}

	// Fall back to object format {buy: [...], sell: [...]}
	var data ordersData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("unexpected orders format: %w", err)
	}
	return append(data.Buy, data.Sell...), nil
}
