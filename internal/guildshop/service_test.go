package guildshop

import (
	"testing"
	"time"
)

func TestSelectGoods(t *testing.T) {
	entries := []StoreEntry{{ID: 1, Weight: 0}, {ID: 2, Weight: 2}}
	selected := selectGoods(entries, 1)
	if len(selected) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(selected))
	}
	if selected[0].ID == 0 {
		t.Fatalf("expected valid entry")
	}
}

func TestSelectGoodsEmpty(t *testing.T) {
	if selectGoods(nil, 1) != nil {
		t.Fatalf("expected nil for empty entries")
	}
}

func TestBuildGoodsDefaults(t *testing.T) {
	config := &Config{StoreEntries: []StoreEntry{{ID: 1, GoodsPurchaseLimit: 0}}, GoodsCount: 1}
	goods := buildGoods(10, config)
	if len(goods) != 1 {
		t.Fatalf("expected 1 good")
	}
	if goods[0].Count != 1 {
		t.Fatalf("expected default count 1")
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
