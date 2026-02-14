package arenashop

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
)

func setupArenaShopTest(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	orm.InitDatabase()
	clearTable(t, "config_entries")
	clearTable(t, "arena_shop_states")
}

func clearTable(t *testing.T, table string) {
	t.Helper()
	query := "DELETE FROM " + table
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), query); err != nil {
		t.Fatalf("failed to clear table: %v", err)
	}
}

func seedArenaShopConfig(t *testing.T, payload string) {
	t.Helper()
	if err := orm.UpsertConfigEntry(arenaShopConfigCategory, "1", json.RawMessage(payload)); err != nil {
		t.Fatalf("seed config entry failed: %v", err)
	}
}

func withNilDefaultStore(t *testing.T) {
	t.Helper()
	originalStore := db.DefaultStore
	db.DefaultStore = nil
	t.Cleanup(func() {
		db.DefaultStore = originalStore
	})
}

func seedCommander(t *testing.T, commanderID uint32) {
	t.Helper()
	if _, err := orm.GetCommanderByAccountID(commanderID); err == nil {
		return
	} else if !db.IsNotFound(err) {
		t.Fatalf("failed to check commander: %v", err)
	}
	if err := orm.CreateCommanderAccountRoot(commanderID, fmt.Sprintf("arena-%d", commanderID), 1, 1); err != nil {
		t.Fatalf("seed commander failed: %v", err)
	}
}

func TestLoadConfigEmpty(t *testing.T) {
	setupArenaShopTest(t)

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if config == nil {
		t.Fatalf("expected config")
	}
	if len(config.Template.RefreshPrice) != 0 || len(config.Template.CommodityList1) != 0 {
		t.Fatalf("expected empty template")
	}
}

func TestLoadConfigSuccess(t *testing.T) {
	setupArenaShopTest(t)
	seedArenaShopConfig(t, `{"commodity_list_1":[[1,2]],"commodity_list_common":[[3,4]],"refresh_price":[10,20]}`)

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(config.Template.CommodityList1) != 1 || config.Template.CommodityList1[0][0] != 1 {
		t.Fatalf("expected commodity list to load")
	}
	if len(config.Template.RefreshPrice) != 2 || config.Template.RefreshPrice[1] != 20 {
		t.Fatalf("expected refresh prices to load")
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	setupArenaShopTest(t)
	seedArenaShopConfig(t, `[]`)

	if _, err := LoadConfig(); err == nil {
		t.Fatalf("expected error for invalid config shape")
	}
}

func TestLoadConfigListError(t *testing.T) {
	withNilDefaultStore(t)

	if _, err := LoadConfig(); err == nil {
		t.Fatalf("expected error from list config entries")
	}
}

func TestEnsureStateCreates(t *testing.T) {
	setupArenaShopTest(t)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	seedCommander(t, 11)

	state, err := EnsureState(11, now)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state.FlashCount != 0 {
		t.Fatalf("expected flash count 0")
	}
	if state.LastRefreshTime != uint32(now.Unix()) {
		t.Fatalf("expected last refresh time to match")
	}
	if state.NextFlashTime != nextDailyReset(now) {
		t.Fatalf("expected next flash time to be next reset")
	}
}

func TestEnsureStateExisting(t *testing.T) {
	setupArenaShopTest(t)
	seedCommander(t, 22)
	seed := orm.ArenaShopState{
		CommanderID:     22,
		FlashCount:      3,
		LastRefreshTime: 10,
		NextFlashTime:   20,
	}
	if err := orm.CreateArenaShopState(seed); err != nil {
		t.Fatalf("seed state failed: %v", err)
	}

	state, err := EnsureState(22, time.Now())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state.FlashCount != 3 || state.LastRefreshTime != 10 || state.NextFlashTime != 20 {
		t.Fatalf("expected existing state to be preserved")
	}
}

func TestEnsureStateError(t *testing.T) {
	withNilDefaultStore(t)

	if _, err := EnsureState(1, time.Now()); err == nil {
		t.Fatalf("expected error from ensure state")
	}
}

func TestEnsureStateCreateError(t *testing.T) {
	withNilDefaultStore(t)

	if _, err := EnsureState(100, time.Now()); err == nil {
		t.Fatalf("expected create error")
	}
}

func TestRefreshIfNeededNoRefresh(t *testing.T) {
	setupArenaShopTest(t)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	seedCommander(t, 30)
	seed := orm.ArenaShopState{
		CommanderID:     30,
		FlashCount:      2,
		LastRefreshTime: uint32(now.Unix()),
		NextFlashTime:   uint32(now.Add(2 * time.Hour).Unix()),
	}
	if err := orm.CreateArenaShopState(seed); err != nil {
		t.Fatalf("seed state failed: %v", err)
	}

	state, err := RefreshIfNeeded(30, now)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state.FlashCount != 2 || state.NextFlashTime != seed.NextFlashTime {
		t.Fatalf("expected state unchanged")
	}
}

func TestRefreshIfNeededResets(t *testing.T) {
	setupArenaShopTest(t)
	now := time.Date(2026, 1, 2, 1, 0, 0, 0, time.UTC)
	seedCommander(t, 31)
	seed := orm.ArenaShopState{
		CommanderID:     31,
		FlashCount:      4,
		LastRefreshTime: 10,
		NextFlashTime:   uint32(now.Add(-1 * time.Hour).Unix()),
	}
	if err := orm.CreateArenaShopState(seed); err != nil {
		t.Fatalf("seed state failed: %v", err)
	}

	state, err := RefreshIfNeeded(31, now)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state.FlashCount != 0 {
		t.Fatalf("expected flash count reset")
	}
	if state.LastRefreshTime != uint32(now.Unix()) {
		t.Fatalf("expected last refresh time updated")
	}
	if state.NextFlashTime != nextDailyReset(now) {
		t.Fatalf("expected next reset computed")
	}
}

func TestRefreshIfNeededEnsureStateError(t *testing.T) {
	withNilDefaultStore(t)

	if _, err := RefreshIfNeeded(40, time.Now()); err == nil {
		t.Fatalf("expected error from refresh if needed")
	}
}

func TestRefreshIfNeededSaveError(t *testing.T) {
	withNilDefaultStore(t)

	if _, err := RefreshIfNeeded(41, time.Now()); err == nil {
		t.Fatalf("expected update error")
	}
}

func TestRefreshShopNilConfig(t *testing.T) {
	setupArenaShopTest(t)
	seedCommander(t, 50)
	state, list, cost, err := RefreshShop(50, time.Now(), nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state == nil || list != nil || cost != 0 {
		t.Fatalf("expected nil list and zero cost")
	}
}

func TestRefreshShopOverRefreshLimit(t *testing.T) {
	setupArenaShopTest(t)
	seedCommander(t, 51)
	seed := orm.ArenaShopState{CommanderID: 51, FlashCount: 1, LastRefreshTime: 10, NextFlashTime: uint32(time.Now().Unix())}
	if err := orm.CreateArenaShopState(seed); err != nil {
		t.Fatalf("seed state failed: %v", err)
	}
	config := &Config{Template: shopTemplate{RefreshPrice: []uint32{5}}}

	state, list, cost, err := RefreshShop(51, time.Now(), config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state.FlashCount != 1 || list != nil || cost != 0 {
		t.Fatalf("expected refresh to be blocked by price limit")
	}
}

func TestRefreshShopSuccess(t *testing.T) {
	setupArenaShopTest(t)
	seedCommander(t, 52)
	config := &Config{Template: shopTemplate{
		CommodityList2:      [][]uint32{{10, 1}},
		CommodityListCommon: [][]uint32{{20, 2}},
		RefreshPrice:        []uint32{7},
	}}

	state, list, cost, err := RefreshShop(52, time.Now(), config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state.FlashCount != 1 {
		t.Fatalf("expected flash count increment")
	}
	if cost != 7 {
		t.Fatalf("expected cost 7, got %d", cost)
	}
	if len(list) != 2 || list[0].GetShopId() != 10 || list[1].GetShopId() != 20 {
		t.Fatalf("expected shop list from tier and common")
	}
}

func TestRefreshShopEnsureStateError(t *testing.T) {
	withNilDefaultStore(t)

	if _, _, _, err := RefreshShop(53, time.Now(), &Config{}); err == nil {
		t.Fatalf("expected error from refresh shop")
	}
}

func TestRefreshShopSaveError(t *testing.T) {
	withNilDefaultStore(t)
	config := &Config{Template: shopTemplate{RefreshPrice: []uint32{3}}}

	if _, _, _, err := RefreshShop(54, time.Now(), config); err == nil {
		t.Fatalf("expected update error")
	}
}

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

func TestBuildArenaShopEmpty(t *testing.T) {
	if list := buildArenaShop(nil); list != nil {
		t.Fatalf("expected nil list for empty entries")
	}
}

func TestBuildShopList(t *testing.T) {
	config := &Config{Template: shopTemplate{
		CommodityList1:      [][]uint32{{1, 1}},
		CommodityList5:      [][]uint32{{5, 5}},
		CommodityListCommon: [][]uint32{{2, 2}},
	}}
	list := BuildShopList(0, config)
	if len(list) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(list))
	}
	list = BuildShopList(4, config)
	if len(list) != 2 || list[0].GetShopId() != 5 {
		t.Fatalf("expected tier 5 entries")
	}
	list = BuildShopList(9, config)
	if len(list) != 1 || list[0].GetShopId() != 2 {
		t.Fatalf("expected common entries for unknown tier")
	}
	if BuildShopList(0, nil) != nil {
		t.Fatalf("expected nil list for nil config")
	}
}

func TestBuildShopListCoversAllTiers(t *testing.T) {
	config := &Config{Template: shopTemplate{
		CommodityList1: [][]uint32{{1, 1}},
		CommodityList2: [][]uint32{{2, 2}},
		CommodityList3: [][]uint32{{3, 3}},
		CommodityList4: [][]uint32{{4, 4}},
		CommodityList5: [][]uint32{{5, 5}},
	}}

	cases := []struct {
		name      string
		flash     uint32
		expected  uint32
		wantCount uint32
	}{
		{name: "tier1", flash: 0, expected: 1, wantCount: 1},
		{name: "tier2", flash: 1, expected: 2, wantCount: 2},
		{name: "tier3", flash: 2, expected: 3, wantCount: 3},
		{name: "tier4", flash: 3, expected: 4, wantCount: 4},
		{name: "tier5", flash: 4, expected: 5, wantCount: 5},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			list := BuildShopList(tc.flash, config)
			if len(list) != 1 {
				t.Fatalf("expected 1 entry, got %d", len(list))
			}
			if list[0].GetShopId() != tc.expected || list[0].GetCount() != tc.wantCount {
				t.Fatalf("unexpected shop entry")
			}
		})
	}
}

func TestBuildShopListUnknownTierWithoutCommon(t *testing.T) {
	config := &Config{Template: shopTemplate{}}
	if list := BuildShopList(99, config); list != nil {
		t.Fatalf("expected nil list")
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
