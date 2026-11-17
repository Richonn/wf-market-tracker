package collector

import (
	"testing"

	"github.com/Richonn/wf-market-tracker/internal/model"
)

func orders(entries ...model.Order) []model.Order { return entries }

func order(typ, status string, plat int) model.Order {
	return model.Order{Type: typ, User: model.OrderUser{Status: status}, Platinum: plat}
}

func TestComputeSnapshot_basic(t *testing.T) {
	s := ComputeSnapshot("frost_prime_set", orders(
		order("sell", "ingame", 100),
		order("sell", "online", 120),
		order("sell", "online", 80),
		order("buy", "ingame", 60),
		order("buy", "online", 50),
	))

	if s.ItemSlug != "frost_prime_set" {
		t.Errorf("slug = %q", s.ItemSlug)
	}
	if s.SellMin != 80 {
		t.Errorf("SellMin = %d, want 80", s.SellMin)
	}
	if s.SellMedian != 100 {
		t.Errorf("SellMedian = %d, want 100", s.SellMedian)
	}
	if s.BuyMax != 60 {
		t.Errorf("BuyMax = %d, want 60", s.BuyMax)
	}
	if s.ActiveSellOrders != 3 {
		t.Errorf("ActiveSellOrders = %d, want 3", s.ActiveSellOrders)
	}
	if s.ActiveBuyOrders != 2 {
		t.Errorf("ActiveBuyOrders = %d, want 2", s.ActiveBuyOrders)
	}
}

func TestComputeSnapshot_filters_offline(t *testing.T) {
	s := ComputeSnapshot("volt_prime_set", orders(
		order("sell", "offline", 10),
		order("sell", "ingame", 200),
		order("buy", "offline", 999),
	))

	if s.SellMin != 200 {
		t.Errorf("SellMin = %d, want 200 (offline filtered)", s.SellMin)
	}
	if s.ActiveSellOrders != 1 {
		t.Errorf("ActiveSellOrders = %d, want 1", s.ActiveSellOrders)
	}
	if s.BuyMax != 0 {
		t.Errorf("BuyMax = %d, want 0 (offline buy filtered)", s.BuyMax)
	}
	if s.ActiveBuyOrders != 0 {
		t.Errorf("ActiveBuyOrders = %d, want 0", s.ActiveBuyOrders)
	}
}

func TestComputeSnapshot_empty_orders(t *testing.T) {
	s := ComputeSnapshot("rhino_prime_set", orders())

	if s.SellMin != 0 || s.SellMedian != 0 || s.BuyMax != 0 {
		t.Errorf("expected zero values for empty orders, got sell_min=%d sell_median=%d buy_max=%d",
			s.SellMin, s.SellMedian, s.BuyMax)
	}
	if s.ActiveSellOrders != 0 || s.ActiveBuyOrders != 0 {
		t.Errorf("expected zero active orders")
	}
}

func TestComputeSnapshot_single_sell(t *testing.T) {
	s := ComputeSnapshot("mag_prime_set", orders(
		order("sell", "online", 55),
	))

	if s.SellMin != 55 {
		t.Errorf("SellMin = %d, want 55", s.SellMin)
	}
	if s.SellMedian != 55 {
		t.Errorf("SellMedian = %d, want 55", s.SellMedian)
	}
}

func TestComputeSnapshot_median_even(t *testing.T) {
	s := ComputeSnapshot("saryn_prime_set", orders(
		order("sell", "online", 10),
		order("sell", "online", 20),
		order("sell", "online", 30),
		order("sell", "online", 40),
	))

	if s.SellMin != 10 {
		t.Errorf("SellMin = %d, want 10", s.SellMin)
	}
	if s.SellMedian != 30 {
		t.Errorf("SellMedian = %d, want 30 (index len/2 = 2)", s.SellMedian)
	}
}
