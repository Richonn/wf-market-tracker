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
	if !item.BulkTradable {
		t.Errorf("BulkTradable = false, want true")
	}

	if resp.Data[1].BulkTradable {
		t.Errorf("cosmetic item should not be bulkTradable")
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

func TestOrdersResponseParsing(t *testing.T) {
	raw := `{
		"data": [
			{"id": "o1", "orderType": "sell", "platinum": 150, "quantity": 1, "userName": "player1", "userStatus": "ingame"},
			{"id": "o2", "orderType": "buy",  "platinum": 100, "quantity": 1, "userName": "player2", "userStatus": "online"},
			{"id": "o3", "orderType": "sell", "platinum": 200, "quantity": 1, "userName": "player3", "userStatus": "offline"}
		]
	}`

	var resp ordersResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(resp.Data) != 3 {
		t.Fatalf("len = %d, want 3", len(resp.Data))
	}

	if resp.Data[0].OrderType != "sell" || resp.Data[0].Platinum != 150 {
		t.Errorf("order[0] = %+v", resp.Data[0])
	}
	if resp.Data[1].UserStatus != "online" {
		t.Errorf("order[1].UserStatus = %q, want online", resp.Data[1].UserStatus)
	}
	if resp.Data[2].UserStatus != "offline" {
		t.Errorf("order[2].UserStatus = %q, want offline", resp.Data[2].UserStatus)
	}
}
