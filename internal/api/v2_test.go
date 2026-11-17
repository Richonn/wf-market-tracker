package api

import (
	"encoding/json"
	"testing"

	"github.com/Richonn/wf-market-tracker/internal/model"
)

func TestItemsResponseParsing(t *testing.T) {
	raw := `{
		"data": [
			{"id": "abc123", "slug": "frost_prime_set", "bulkTradable": true, "tags": ["prime", "set"],
			 "i18n": {"en": {"name": "Frost Prime Set"}}},
			{"id": "def456", "slug": "some_cosmetic", "bulkTradable": false, "tags": ["cosmetic"],
			 "i18n": {"en": {"name": "Some Cosmetic"}}}
		]
	}`

	var resp itemsResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(resp.Data) != 2 {
		t.Fatalf("len = %d, want 2", len(resp.Data))
	}

	item := resp.Data[0]
	if item.ID != "abc123" {
		t.Errorf("ID = %q", item.ID)
	}
	if item.Slug != "frost_prime_set" {
		t.Errorf("Slug = %q", item.Slug)
	}
	if item.Name() != "Frost Prime Set" {
		t.Errorf("Name() = %q, want %q", item.Name(), "Frost Prime Set")
	}
}

func TestItemsResponseParsing_filterTradable(t *testing.T) {
	raw := `{
		"data": [
			{"id": "1", "slug": "a", "bulkTradable": true},
			{"id": "2", "slug": "b", "bulkTradable": false},
			{"id": "3", "slug": "c", "bulkTradable": true}
		]
	}`

	var resp itemsResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	var tradeable []model.Item
	for _, item := range resp.Data {
		if item.BulkTradable {
			tradeable = append(tradeable, item)
		}
	}

	if len(tradeable) != 2 {
		t.Errorf("tradeable count = %d, want 2", len(tradeable))
	}
}

func TestOrdersResponseParsing_array(t *testing.T) {
	raw := `{
		"data": [
			{"id": "o1", "type": "sell", "platinum": 150, "quantity": 1, "user": {"status": "ingame"}},
			{"id": "o2", "type": "buy",  "platinum": 100, "quantity": 1, "user": {"status": "online"}},
			{"id": "o3", "type": "sell", "platinum": 200, "quantity": 1, "user": {"status": "offline"}}
		]
	}`

	var resp ordersResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	var orders []model.Order
	if err := json.Unmarshal(resp.Data, &orders); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}

	if len(orders) != 3 {
		t.Fatalf("len = %d, want 3", len(orders))
	}

	if orders[0].Type != "sell" || orders[0].Platinum != 150 {
		t.Errorf("order[0] = %+v", orders[0])
	}
	if orders[1].User.Status != "online" {
		t.Errorf("order[1].User.Status = %q, want online", orders[1].User.Status)
	}
	if orders[2].User.Status != "offline" {
		t.Errorf("order[2].User.Status = %q, want offline", orders[2].User.Status)
	}
}

func TestOrdersResponseParsing_object(t *testing.T) {
	raw := `{
		"data": {
			"sell": [{"id": "o1", "type": "sell", "platinum": 150, "user": {"status": "ingame"}}],
			"buy":  [{"id": "o2", "type": "buy",  "platinum": 100, "user": {"status": "online"}}]
		}
	}`

	var resp ordersResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	var data ordersData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}

	if len(data.Sell) != 1 || len(data.Buy) != 1 {
		t.Fatalf("sell=%d buy=%d, want 1 each", len(data.Sell), len(data.Buy))
	}
	if data.Sell[0].Platinum != 150 {
		t.Errorf("sell[0].Platinum = %d, want 150", data.Sell[0].Platinum)
	}
}
