package arenashop

import (
	"testing"
	"time"
)

func TestBuildArenaShop(t *testing.T) {
	entries := [][]uint32{{1, 2}, {3}}
	list := buildArenaShop(entries)
	if len(list) != 1 {
		t.Fatalf("expected 1 valid entry, got %d", len(list))
	}
	if list[0].GetShopId() != 1 || list[0].GetCount() != 2 {
		t.Fatalf("unexpected shop entry")
	}
}

func TestBuildShopList(t *testing.T) {
	config := &Config{Template: shopTemplate{
		CommodityList1:      [][]uint32{{1, 1}},
		CommodityListCommon: [][]uint32{{2, 2}},
	}}
	list := BuildShopList(0, config)
	if len(list) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(list))
	}
}

func TestNextDailyReset(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	reset := nextDailyReset(now)
	expected := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	if reset != uint32(expected.Unix()) {
		t.Fatalf("expected next daily reset")
	}
}
