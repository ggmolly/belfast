package minigameshop

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
)

func setupMiniGameShopTest(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	orm.InitDatabase()
	clearMiniGameShopTables(t, "config_entries", "mini_game_shop_states", "mini_game_shop_goods")
}

func clearMiniGameShopTables(t *testing.T, tables ...string) {
	t.Helper()
	if len(tables) == 0 {
		return
	}
	query := "TRUNCATE TABLE " + strings.Join(tables, ", ") + " RESTART IDENTITY CASCADE"
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), query); err != nil {
		t.Fatalf("failed to clear tables: %v", err)
	}
}

func withNilDefaultStore(t *testing.T, fn func()) {
	t.Helper()
	originalStore := db.DefaultStore
	db.DefaultStore = nil
	defer func() {
		db.DefaultStore = originalStore
	}()
	fn()
}

func seedConfigEntry(t *testing.T, category, key, payload string) {
	t.Helper()
	entry := orm.ConfigEntry{
		Category: category,
		Key:      key,
		Data:     json.RawMessage(payload),
	}
	if err := orm.UpsertConfigEntry(entry.Category, entry.Key, entry.Data); err != nil {
		t.Fatalf("seed config entry failed: %v", err)
	}
}

func seedCommander(t *testing.T, commanderID uint32) {
	t.Helper()
	if err := orm.CreateCommanderRoot(commanderID, commanderID, "MiniGameShop Tester", 0, 0); err != nil {
		t.Fatalf("seed commander failed: %v", err)
	}
}

func TestLoadConfigFiltersAndSorts(t *testing.T) {
	setupMiniGameShopTest(t)
	now := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	within := [][][3]int{{{2026, 1, 1}, {2026, 1, 3}}}
	outside := [][][3]int{{{2025, 1, 1}, {2025, 1, 2}}}
	payload, err := json.Marshal(shopEntry{ID: 0, GoodsPurchaseLimit: 1, Order: 1})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	seedConfigEntry(t, gameRoomShopCategory, "1", string(payload))
	for key, entry := range map[string]shopEntry{
		"20": {ID: 2, GoodsPurchaseLimit: 4, Order: 2, Time: within},
		"10": {ID: 1, GoodsPurchaseLimit: 2, Order: 1, Time: within},
		"30": {ID: 3, GoodsPurchaseLimit: 1, Order: 1, Time: outside},
	} {
		payload, err := json.Marshal(entry)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		seedConfigEntry(t, gameRoomShopCategory, key, string(payload))
	}

	config, err := LoadConfig(now)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if config == nil {
		t.Fatalf("expected config")
	}
	if len(config.Goods) != 2 {
		t.Fatalf("expected 2 goods, got %d", len(config.Goods))
	}
	if config.Goods[0].ID != 1 || config.Goods[1].ID != 2 {
		t.Fatalf("expected sorted goods")
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	setupMiniGameShopTest(t)
	seedConfigEntry(t, gameRoomShopCategory, "bad", `{"id":"bad"}`)

	if _, err := LoadConfig(time.Now()); err == nil {
		t.Fatalf("expected error for invalid json")
	}
}

func TestLoadConfigListError(t *testing.T) {
	withNilDefaultStore(t, func() {
		if _, err := LoadConfig(time.Now()); err == nil {
			t.Fatalf("expected error from list config entries")
		}
	})
}

func TestEnsureStateCreates(t *testing.T) {
	setupMiniGameShopTest(t)
	seedCommander(t, 10)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	config := &Config{Goods: []shopEntry{{ID: 1, GoodsPurchaseLimit: 0}, {ID: 2, GoodsPurchaseLimit: 3}}}

	state, goods, err := EnsureState(10, now, config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state.NextRefreshTime != nextDailyReset(now) {
		t.Fatalf("expected next refresh time set")
	}
	if len(goods) != 2 {
		t.Fatalf("expected 2 goods, got %d", len(goods))
	}
	if goods[0].CommanderID != 10 || goods[0].GoodsID != 1 || goods[0].Count != 1 {
		t.Fatalf("unexpected first good")
	}
	if goods[1].GoodsID != 2 || goods[1].Count != 3 {
		t.Fatalf("unexpected second good")
	}
}

func TestEnsureStateExisting(t *testing.T) {
	setupMiniGameShopTest(t)
	seedCommander(t, 20)
	seed := orm.MiniGameShopState{CommanderID: 20, NextRefreshTime: 99}
	if err := orm.CreateMiniGameShopState(seed); err != nil {
		t.Fatalf("seed state failed: %v", err)
	}
	seedGoods := []orm.MiniGameShopGood{{CommanderID: 20, GoodsID: 11, Count: 1}}
	for i := range seedGoods {
		if err := orm.CreateMiniGameShopGood(seedGoods[i]); err != nil {
			t.Fatalf("seed goods failed: %v", err)
		}
	}

	state, goods, err := EnsureState(20, time.Now(), &Config{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state.NextRefreshTime != 99 {
		t.Fatalf("expected existing state")
	}
	if len(goods) != 1 || goods[0].GoodsID != 11 {
		t.Fatalf("expected existing goods")
	}
}

func TestEnsureStateError(t *testing.T) {
	withNilDefaultStore(t, func() {
		if _, _, err := EnsureState(1, time.Now(), &Config{}); err == nil {
			t.Fatalf("expected error from ensure state")
		}
	})
}

func TestRefreshIfNeededNoRefresh(t *testing.T) {
	setupMiniGameShopTest(t)
	seedCommander(t, 30)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	seed := orm.MiniGameShopState{CommanderID: 30, NextRefreshTime: uint32(now.Add(2 * time.Hour).Unix())}
	if err := orm.CreateMiniGameShopState(seed); err != nil {
		t.Fatalf("seed state failed: %v", err)
	}
	seedGoods := []orm.MiniGameShopGood{{CommanderID: 30, GoodsID: 100, Count: 1}}
	for i := range seedGoods {
		if err := orm.CreateMiniGameShopGood(seedGoods[i]); err != nil {
			t.Fatalf("seed goods failed: %v", err)
		}
	}
	config := &Config{Goods: []shopEntry{{ID: 5, GoodsPurchaseLimit: 2}}}

	state, goods, err := RefreshIfNeeded(30, now, config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state.NextRefreshTime != seed.NextRefreshTime {
		t.Fatalf("expected state unchanged")
	}
	if len(goods) != 1 || goods[0].GoodsID != 100 {
		t.Fatalf("expected goods unchanged")
	}
}

func TestRefreshIfNeededRefreshesOnTime(t *testing.T) {
	setupMiniGameShopTest(t)
	seedCommander(t, 31)
	now := time.Date(2026, 1, 2, 1, 0, 0, 0, time.UTC)
	seed := orm.MiniGameShopState{CommanderID: 31, NextRefreshTime: uint32(now.Add(-1 * time.Hour).Unix())}
	if err := orm.CreateMiniGameShopState(seed); err != nil {
		t.Fatalf("seed state failed: %v", err)
	}
	seedGoods := []orm.MiniGameShopGood{{CommanderID: 31, GoodsID: 200, Count: 1}}
	for i := range seedGoods {
		if err := orm.CreateMiniGameShopGood(seedGoods[i]); err != nil {
			t.Fatalf("seed goods failed: %v", err)
		}
	}
	config := &Config{Goods: []shopEntry{{ID: 7, GoodsPurchaseLimit: 1}}}

	state, goods, err := RefreshIfNeeded(31, now, config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state.NextRefreshTime != nextDailyReset(now) {
		t.Fatalf("expected next refresh time updated")
	}
	if len(goods) != 1 || goods[0].GoodsID != 7 {
		t.Fatalf("expected refreshed goods")
	}
}

func TestRefreshIfNeededRefreshesOnEmptyGoods(t *testing.T) {
	setupMiniGameShopTest(t)
	seedCommander(t, 32)
	now := time.Date(2026, 1, 3, 10, 0, 0, 0, time.UTC)
	seed := orm.MiniGameShopState{CommanderID: 32, NextRefreshTime: uint32(now.Add(2 * time.Hour).Unix())}
	if err := orm.CreateMiniGameShopState(seed); err != nil {
		t.Fatalf("seed state failed: %v", err)
	}
	config := &Config{Goods: []shopEntry{{ID: 9, GoodsPurchaseLimit: 2}}}

	state, goods, err := RefreshIfNeeded(32, now, config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state.NextRefreshTime != nextDailyReset(now) {
		t.Fatalf("expected next refresh time updated")
	}
	if len(goods) != 1 || goods[0].GoodsID != 9 {
		t.Fatalf("expected goods refreshed when empty")
	}
}

func TestRefreshGoodsSuccess(t *testing.T) {
	setupMiniGameShopTest(t)
	seedCommander(t, 40)
	seed := orm.MiniGameShopState{CommanderID: 40, NextRefreshTime: 10}
	if err := orm.CreateMiniGameShopState(seed); err != nil {
		t.Fatalf("seed state failed: %v", err)
	}
	seedGoods := []orm.MiniGameShopGood{{CommanderID: 40, GoodsID: 100, Count: 1}}
	for i := range seedGoods {
		if err := orm.CreateMiniGameShopGood(seedGoods[i]); err != nil {
			t.Fatalf("seed goods failed: %v", err)
		}
	}
	config := &Config{Goods: []shopEntry{{ID: 5, GoodsPurchaseLimit: 4}}}

	goods, err := RefreshGoods(40, config, RefreshOptions{NextRefreshTime: 77})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(goods) != 1 || goods[0].GoodsID != 5 || goods[0].Count != 4 {
		t.Fatalf("expected refreshed goods")
	}
	state, err := orm.GetMiniGameShopState(40)
	if err != nil {
		t.Fatalf("expected state, got %v", err)
	}
	if state.NextRefreshTime != 77 {
		t.Fatalf("expected state updated")
	}
}

func TestRefreshGoodsNilConfigDeletes(t *testing.T) {
	setupMiniGameShopTest(t)
	seedCommander(t, 41)
	seed := orm.MiniGameShopState{CommanderID: 41, NextRefreshTime: 10}
	if err := orm.CreateMiniGameShopState(seed); err != nil {
		t.Fatalf("seed state failed: %v", err)
	}
	seedGoods := []orm.MiniGameShopGood{{CommanderID: 41, GoodsID: 100, Count: 1}}
	for i := range seedGoods {
		if err := orm.CreateMiniGameShopGood(seedGoods[i]); err != nil {
			t.Fatalf("seed goods failed: %v", err)
		}
	}

	goods, err := RefreshGoods(41, nil, RefreshOptions{NextRefreshTime: 55})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if goods != nil {
		t.Fatalf("expected nil goods for nil config")
	}
	remaining, err := orm.LoadMiniGameShopGoods(41)
	if err != nil {
		t.Fatalf("expected goods query, got %v", err)
	}
	if len(remaining) != 0 {
		t.Fatalf("expected goods deleted")
	}
}

func TestRefreshGoodsRollbackOnUpdateError(t *testing.T) {
	setupMiniGameShopTest(t)
	seedCommander(t, 42)
	seed := orm.MiniGameShopState{CommanderID: 42, NextRefreshTime: 44}
	if err := orm.CreateMiniGameShopState(seed); err != nil {
		t.Fatalf("seed state failed: %v", err)
	}
	seedGoods := []orm.MiniGameShopGood{{CommanderID: 42, GoodsID: 200, Count: 2}}
	for i := range seedGoods {
		if err := orm.CreateMiniGameShopGood(seedGoods[i]); err != nil {
			t.Fatalf("seed goods failed: %v", err)
		}
	}
	withNilDefaultStore(t, func() {
		config := &Config{Goods: []shopEntry{{ID: 9, GoodsPurchaseLimit: 1}}}

		if _, err := RefreshGoods(42, config, RefreshOptions{NextRefreshTime: 99}); err == nil {
			t.Fatalf("expected update error")
		}
	})
	goods, err := orm.LoadMiniGameShopGoods(42)
	if err != nil {
		t.Fatalf("expected goods query, got %v", err)
	}
	if len(goods) != 1 || goods[0].GoodsID != 200 {
		t.Fatalf("expected goods unchanged after rollback")
	}
	state, err := orm.GetMiniGameShopState(42)
	if err != nil {
		t.Fatalf("expected state query, got %v", err)
	}
	if state.NextRefreshTime != 44 {
		t.Fatalf("expected state unchanged after rollback")
	}
}

func TestLoadGoods(t *testing.T) {
	setupMiniGameShopTest(t)
	seedCommander(t, 50)
	seedGoods := []orm.MiniGameShopGood{{CommanderID: 50, GoodsID: 10, Count: 1}, {CommanderID: 50, GoodsID: 11, Count: 2}}
	for i := range seedGoods {
		if err := orm.CreateMiniGameShopGood(seedGoods[i]); err != nil {
			t.Fatalf("seed goods failed: %v", err)
		}
	}

	goods, err := LoadGoods(50)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(goods) != 2 {
		t.Fatalf("expected 2 goods, got %d", len(goods))
	}
}

func TestLoadGoodsError(t *testing.T) {
	withNilDefaultStore(t, func() {
		if _, err := LoadGoods(1); err == nil {
			t.Fatalf("expected error from load goods")
		}
	})
}

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
	if goods[0].CommanderID != 10 || goods[0].GoodsID != 1 {
		t.Fatalf("expected commander and goods id")
	}
}

func TestBuildGoodsNilConfig(t *testing.T) {
	if goods := buildGoods(1, nil); goods != nil {
		t.Fatalf("expected nil goods for nil config")
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
	if isWithinTime(now, [][][3]int{{{0, 0, 0}, {2026, 1, 3}}}) {
		t.Fatalf("expected false for zero start")
	}
	window := [][][3]int{{{2026, 1, 1}, {2026, 1, 3}}}
	if !isWithinTime(now, window) {
		t.Fatalf("expected true within window")
	}
	if isWithinTime(now, [][][3]int{{{2026, 1, 3}, {2026, 1, 4}}}) {
		t.Fatalf("expected false outside window")
	}
	if !isWithinTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), window) {
		t.Fatalf("expected inclusive start")
	}
	if !isWithinTime(time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC), window) {
		t.Fatalf("expected inclusive end")
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
