package arenashop

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupArenaShopTest(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	orm.InitDatabase()
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.ArenaShopState{})
}

func clearTable(t *testing.T, model any) {
	t.Helper()
	if err := orm.GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(model).Error; err != nil {
		t.Fatalf("failed to clear table: %v", err)
	}
}

func seedArenaShopConfig(t *testing.T, payload string) {
	t.Helper()
	entry := orm.ConfigEntry{
		Category: arenaShopConfigCategory,
		Key:      "1",
		Data:     json.RawMessage(payload),
	}
	if err := orm.GormDB.Create(&entry).Error; err != nil {
		t.Fatalf("seed config entry failed: %v", err)
	}
}

func newTestDB(t *testing.T, models ...any) *gorm.DB {
	t.Helper()
	name := strings.ReplaceAll(t.Name(), "/", "_")
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", name)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{PrepareStmt: true})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	if len(models) > 0 {
		if err := db.AutoMigrate(models...); err != nil {
			t.Fatalf("failed to migrate: %v", err)
		}
	}
	return db
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
	seedArenaShopConfig(t, `{"commodity_list_1":`) // invalid JSON

	if _, err := LoadConfig(); err == nil {
		t.Fatalf("expected error for invalid json")
	}
}

func TestLoadConfigListError(t *testing.T) {
	originalDB := orm.GormDB
	defer func() {
		orm.GormDB = originalDB
	}()
	orm.GormDB = newTestDB(t)

	if _, err := LoadConfig(); err == nil {
		t.Fatalf("expected error from list config entries")
	}
}

func TestEnsureStateCreates(t *testing.T) {
	setupArenaShopTest(t)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

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
	seed := orm.ArenaShopState{
		CommanderID:     22,
		FlashCount:      3,
		LastRefreshTime: 10,
		NextFlashTime:   20,
	}
	if err := orm.GormDB.Create(&seed).Error; err != nil {
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
	originalDB := orm.GormDB
	defer func() {
		orm.GormDB = originalDB
	}()
	orm.GormDB = newTestDB(t)

	if _, err := EnsureState(1, time.Now()); err == nil {
		t.Fatalf("expected error from ensure state")
	}
}

func TestRefreshIfNeededNoRefresh(t *testing.T) {
	setupArenaShopTest(t)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	seed := orm.ArenaShopState{
		CommanderID:     30,
		FlashCount:      2,
		LastRefreshTime: uint32(now.Unix()),
		NextFlashTime:   uint32(now.Add(2 * time.Hour).Unix()),
	}
	if err := orm.GormDB.Create(&seed).Error; err != nil {
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
	seed := orm.ArenaShopState{
		CommanderID:     31,
		FlashCount:      4,
		LastRefreshTime: 10,
		NextFlashTime:   uint32(now.Add(-1 * time.Hour).Unix()),
	}
	if err := orm.GormDB.Create(&seed).Error; err != nil {
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
	originalDB := orm.GormDB
	defer func() {
		orm.GormDB = originalDB
	}()
	orm.GormDB = newTestDB(t)

	if _, err := RefreshIfNeeded(40, time.Now()); err == nil {
		t.Fatalf("expected error from refresh if needed")
	}
}

func TestRefreshIfNeededSaveError(t *testing.T) {
	originalDB := orm.GormDB
	defer func() {
		orm.GormDB = originalDB
	}()
	db := newTestDB(t, &orm.ArenaShopState{})
	orm.GormDB = db

	seed := orm.ArenaShopState{
		CommanderID:     41,
		FlashCount:      1,
		LastRefreshTime: 10,
		NextFlashTime:   uint32(time.Now().Add(-1 * time.Hour).Unix()),
	}
	if err := orm.GormDB.Create(&seed).Error; err != nil {
		t.Fatalf("seed state failed: %v", err)
	}

	orm.GormDB.Callback().Update().Replace("gorm:update", func(tx *gorm.DB) {
		tx.AddError(errors.New("update failed"))
	})

	if _, err := RefreshIfNeeded(41, time.Now()); err == nil {
		t.Fatalf("expected update error")
	}
}

func TestRefreshShopNilConfig(t *testing.T) {
	setupArenaShopTest(t)
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
	seed := orm.ArenaShopState{CommanderID: 51, FlashCount: 1, LastRefreshTime: 10, NextFlashTime: uint32(time.Now().Unix())}
	if err := orm.GormDB.Create(&seed).Error; err != nil {
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
	originalDB := orm.GormDB
	defer func() {
		orm.GormDB = originalDB
	}()
	orm.GormDB = newTestDB(t)

	if _, _, _, err := RefreshShop(53, time.Now(), &Config{}); err == nil {
		t.Fatalf("expected error from refresh shop")
	}
}

func TestRefreshShopSaveError(t *testing.T) {
	originalDB := orm.GormDB
	defer func() {
		orm.GormDB = originalDB
	}()
	db := newTestDB(t, &orm.ArenaShopState{})
	orm.GormDB = db
	seed := orm.ArenaShopState{CommanderID: 54, FlashCount: 0, LastRefreshTime: 10, NextFlashTime: uint32(time.Now().Unix())}
	if err := orm.GormDB.Create(&seed).Error; err != nil {
		t.Fatalf("seed state failed: %v", err)
	}
	orm.GormDB.Callback().Update().Replace("gorm:update", func(tx *gorm.DB) {
		tx.AddError(errors.New("update failed"))
	})
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

func TestNextDailyReset(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	reset := nextDailyReset(now)
	expected := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	if reset != uint32(expected.Unix()) {
		t.Fatalf("expected next daily reset")
	}
}
