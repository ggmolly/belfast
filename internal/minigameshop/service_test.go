package minigameshop

import (
	"testing"
	"time"
)

func TestBuildGoods(t *testing.T) {
	config := &Config{Goods: []shopEntry{{ID: 1, GoodsPurchaseLimit: 0}, {ID: 2, GoodsPurchaseLimit: 3}}}
	goods := buildGoods(10, config)
	if len(goods) != 2 {
		t.Fatalf("expected 2 goods")
	}
	if goods[0].Count != 1 {
		t.Fatalf("expected default count 1")
	}
	if goods[1].Count != 3 {
		t.Fatalf("expected configured count 3")
	}
}

func TestIsWithinTime(t *testing.T) {
	now := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	if !isWithinTime(now, nil) {
		t.Fatalf("expected true for empty ranges")
	}
	if isWithinTime(now, [][][3]int{{{2025, 1, 1}}}) {
		t.Fatalf("expected false for invalid window")
	}
	window := [][][3]int{{{2026, 1, 1}, {2026, 1, 3}}}
	if !isWithinTime(now, window) {
		t.Fatalf("expected true within window")
	}
}

func TestTimeFromConfig(t *testing.T) {
	if !timeFromConfig(time.UTC, [3]int{0, 0, 0}).IsZero() {
		t.Fatalf("expected zero time")
	}
	result := timeFromConfig(time.UTC, [3]int{2026, 1, 2})
	if result.Year() != 2026 || result.Month() != time.January || result.Day() != 2 {
		t.Fatalf("unexpected time %v", result)
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
