package guildshop

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

func withTestDB(t *testing.T, models ...any) {
	t.Helper()
	originalDB := orm.GormDB
	orm.GormDB = newTestDB(t, models...)
	t.Cleanup(func() {
		orm.GormDB = originalDB
	})
}

func seedConfigEntry(t *testing.T, category, key, payload string) {
	t.Helper()
	entry := orm.ConfigEntry{
		Category: category,
		Key:      key,
		Data:     json.RawMessage(payload),
	}
	if err := orm.GormDB.Create(&entry).Error; err != nil {
		t.Fatalf("seed config entry failed: %v", err)
	}
}

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

func TestSelectGoodsWeightBias(t *testing.T) {
	entries := []StoreEntry{{ID: 1, Weight: 1}, {ID: 2, Weight: 1000}}
	selectedHigh := 0
	for i := 0; i < 200; i++ {
		selected := selectGoods(entries, 1)
		if len(selected) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(selected))
		}
		if selected[0].ID == 2 {
			selectedHigh++
		}
	}
	if selectedHigh < 190 {
		t.Fatalf("expected weighted selection to favor high weight, got %d", selectedHigh)
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

func TestBuildGoodsNilConfig(t *testing.T) {
	if goods := buildGoods(1, nil); goods != nil {
		t.Fatalf("expected nil goods for nil config")
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	withTestDB(t, &orm.ConfigEntry{})
	seedConfigEntry(t, guildStoreConfigCategory, "1", `{"id":0,"weight":1,"goods_purchase_limit":1}`)
	seedConfigEntry(t, guildStoreConfigCategory, "2", `{"id":123,"weight":2,"goods_purchase_limit":0}`)
	seedConfigEntry(t, guildSetConfigCategory, "store_goods_quantity", `{"key":"store_goods_quantity","key_value":0,"key_args":[1]}`)
	seedConfigEntry(t, guildSetConfigCategory, "store_reset_cost", `{"key":"store_reset_cost","key_value":15,"key_args":[2]}`)

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if config == nil {
		t.Fatalf("expected config")
	}
	if config.GoodsCount != 10 {
		t.Fatalf("expected default goods count 10, got %d", config.GoodsCount)
	}
	if config.ResetCost != 15 {
		t.Fatalf("expected reset cost 15, got %d", config.ResetCost)
	}
	if len(config.StoreEntries) != 1 || config.StoreEntries[0].ID != 123 {
		t.Fatalf("expected filtered store entries")
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	withTestDB(t, &orm.ConfigEntry{})
	seedConfigEntry(t, guildStoreConfigCategory, "1", `{"id":`) // invalid JSON

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

func TestGetGuildSetValue(t *testing.T) {
	withTestDB(t, &orm.ConfigEntry{})
	seedConfigEntry(t, guildSetConfigCategory, "store_goods_quantity", `{"key":"store_goods_quantity","key_value":7}`)

	value, err := getGuildSetValue("store_goods_quantity")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if value != 7 {
		t.Fatalf("expected value 7, got %d", value)
	}
}

func TestGetGuildSetValueInvalidJSON(t *testing.T) {
	withTestDB(t, &orm.ConfigEntry{})
	seedConfigEntry(t, guildSetConfigCategory, "store_goods_quantity", `{"key":`) // invalid JSON

	if _, err := getGuildSetValue("store_goods_quantity"); err == nil {
		t.Fatalf("expected error for invalid json")
	}
}

func TestEnsureStateCreates(t *testing.T) {
	withTestDB(t, &orm.GuildShopState{}, &orm.GuildShopGood{})
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	config := &Config{StoreEntries: []StoreEntry{{ID: 1, GoodsPurchaseLimit: 2}, {ID: 2, GoodsPurchaseLimit: 3}}, GoodsCount: 5}

	state, goods, err := EnsureState(10, now, config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state.RefreshCount != 0 {
		t.Fatalf("expected refresh count 0")
	}
	if state.NextRefreshTime != nextDailyReset(now) {
		t.Fatalf("expected next refresh time to be next reset")
	}
	if len(goods) != 2 {
		t.Fatalf("expected 2 goods, got %d", len(goods))
	}
	if goods[0].CommanderID != 10 || goods[0].Index != 1 || goods[0].GoodsID != 1 || goods[0].Count != 2 {
		t.Fatalf("unexpected first good")
	}
	if goods[1].Index != 2 || goods[1].GoodsID != 2 || goods[1].Count != 3 {
		t.Fatalf("unexpected second good")
	}
}

func TestEnsureStateExisting(t *testing.T) {
	withTestDB(t, &orm.GuildShopState{}, &orm.GuildShopGood{})
	seed := orm.GuildShopState{CommanderID: 20, RefreshCount: 3, NextRefreshTime: 99}
	if err := orm.GormDB.Create(&seed).Error; err != nil {
		t.Fatalf("seed state failed: %v", err)
	}
	goodsSeed := []orm.GuildShopGood{{CommanderID: 20, Index: 1, GoodsID: 11, Count: 1}}
	if err := orm.GormDB.Create(&goodsSeed).Error; err != nil {
		t.Fatalf("seed goods failed: %v", err)
	}

	state, goods, err := EnsureState(20, time.Now(), &Config{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state.RefreshCount != 3 || state.NextRefreshTime != 99 {
		t.Fatalf("expected existing state")
	}
	if len(goods) != 1 || goods[0].GoodsID != 11 {
		t.Fatalf("expected existing goods")
	}
}

func TestEnsureStateError(t *testing.T) {
	originalDB := orm.GormDB
	defer func() {
		orm.GormDB = originalDB
	}()
	orm.GormDB = newTestDB(t)

	if _, _, err := EnsureState(1, time.Now(), &Config{}); err == nil {
		t.Fatalf("expected error from ensure state")
	}
}

func TestRefreshIfNeededNoRefresh(t *testing.T) {
	withTestDB(t, &orm.GuildShopState{}, &orm.GuildShopGood{})
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	seed := orm.GuildShopState{CommanderID: 30, RefreshCount: 2, NextRefreshTime: uint32(now.Add(2 * time.Hour).Unix())}
	if err := orm.GormDB.Create(&seed).Error; err != nil {
		t.Fatalf("seed state failed: %v", err)
	}
	seedGoods := []orm.GuildShopGood{{CommanderID: 30, Index: 1, GoodsID: 100, Count: 1}}
	if err := orm.GormDB.Create(&seedGoods).Error; err != nil {
		t.Fatalf("seed goods failed: %v", err)
	}

	state, goods, err := RefreshIfNeeded(30, now, &Config{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state.RefreshCount != 2 || state.NextRefreshTime != seed.NextRefreshTime {
		t.Fatalf("expected state unchanged")
	}
	if len(goods) != 1 || goods[0].GoodsID != 100 {
		t.Fatalf("expected goods unchanged")
	}
}

func TestRefreshIfNeededRefreshesOnTime(t *testing.T) {
	withTestDB(t, &orm.GuildShopState{}, &orm.GuildShopGood{})
	now := time.Date(2026, 1, 2, 1, 0, 0, 0, time.UTC)
	seed := orm.GuildShopState{CommanderID: 31, RefreshCount: 4, NextRefreshTime: uint32(now.Add(-1 * time.Hour).Unix())}
	if err := orm.GormDB.Create(&seed).Error; err != nil {
		t.Fatalf("seed state failed: %v", err)
	}
	seedGoods := []orm.GuildShopGood{{CommanderID: 31, Index: 1, GoodsID: 200, Count: 1}}
	if err := orm.GormDB.Create(&seedGoods).Error; err != nil {
		t.Fatalf("seed goods failed: %v", err)
	}
	config := &Config{StoreEntries: []StoreEntry{{ID: 7, GoodsPurchaseLimit: 1}}, GoodsCount: 1}

	state, goods, err := RefreshIfNeeded(31, now, config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state.RefreshCount != 0 {
		t.Fatalf("expected refresh count reset")
	}
	if state.NextRefreshTime != nextDailyReset(now) {
		t.Fatalf("expected next refresh time updated")
	}
	if len(goods) != 1 || goods[0].GoodsID != 7 {
		t.Fatalf("expected refreshed goods")
	}
}

func TestRefreshIfNeededRefreshesOnEmptyGoods(t *testing.T) {
	withTestDB(t, &orm.GuildShopState{}, &orm.GuildShopGood{})
	now := time.Date(2026, 1, 3, 10, 0, 0, 0, time.UTC)
	seed := orm.GuildShopState{CommanderID: 32, RefreshCount: 1, NextRefreshTime: uint32(now.Add(2 * time.Hour).Unix())}
	if err := orm.GormDB.Create(&seed).Error; err != nil {
		t.Fatalf("seed state failed: %v", err)
	}
	config := &Config{StoreEntries: []StoreEntry{{ID: 9, GoodsPurchaseLimit: 2}}, GoodsCount: 1}

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
	withTestDB(t, &orm.GuildShopState{}, &orm.GuildShopGood{})
	seed := orm.GuildShopState{CommanderID: 40, RefreshCount: 2, NextRefreshTime: 10}
	if err := orm.GormDB.Create(&seed).Error; err != nil {
		t.Fatalf("seed state failed: %v", err)
	}
	seedGoods := []orm.GuildShopGood{{CommanderID: 40, Index: 1, GoodsID: 100, Count: 1}}
	if err := orm.GormDB.Create(&seedGoods).Error; err != nil {
		t.Fatalf("seed goods failed: %v", err)
	}
	config := &Config{StoreEntries: []StoreEntry{{ID: 5, GoodsPurchaseLimit: 4}}, GoodsCount: 1}

	goods, err := RefreshGoods(40, time.Now(), config, RefreshOptions{RefreshCount: 0, NextRefreshTime: 77})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(goods) != 1 || goods[0].GoodsID != 5 || goods[0].Count != 4 {
		t.Fatalf("expected refreshed goods")
	}
	var state orm.GuildShopState
	if err := orm.GormDB.Where("commander_id = ?", 40).First(&state).Error; err != nil {
		t.Fatalf("expected state, got %v", err)
	}
	if state.RefreshCount != 0 || state.NextRefreshTime != 77 {
		t.Fatalf("expected state updated")
	}
}

func TestRefreshGoodsRollbackOnUpdateError(t *testing.T) {
	withTestDB(t, &orm.GuildShopState{}, &orm.GuildShopGood{})
	seed := orm.GuildShopState{CommanderID: 41, RefreshCount: 3, NextRefreshTime: 44}
	if err := orm.GormDB.Create(&seed).Error; err != nil {
		t.Fatalf("seed state failed: %v", err)
	}
	seedGoods := []orm.GuildShopGood{{CommanderID: 41, Index: 1, GoodsID: 200, Count: 2}}
	if err := orm.GormDB.Create(&seedGoods).Error; err != nil {
		t.Fatalf("seed goods failed: %v", err)
	}
	orm.GormDB.Callback().Update().Replace("gorm:update", func(tx *gorm.DB) {
		tx.AddError(errors.New("update failed"))
	})
	config := &Config{StoreEntries: []StoreEntry{{ID: 9, GoodsPurchaseLimit: 1}}, GoodsCount: 1}

	if _, err := RefreshGoods(41, time.Now(), config, RefreshOptions{RefreshCount: 0, NextRefreshTime: 99}); err == nil {
		t.Fatalf("expected update error")
	}
	var goods []orm.GuildShopGood
	if err := orm.GormDB.Where("commander_id = ?", 41).Find(&goods).Error; err != nil {
		t.Fatalf("expected goods query, got %v", err)
	}
	if len(goods) != 1 || goods[0].GoodsID != 200 {
		t.Fatalf("expected goods unchanged after rollback")
	}
	var state orm.GuildShopState
	if err := orm.GormDB.Where("commander_id = ?", 41).First(&state).Error; err != nil {
		t.Fatalf("expected state query, got %v", err)
	}
	if state.RefreshCount != 3 || state.NextRefreshTime != 44 {
		t.Fatalf("expected state unchanged after rollback")
	}
}

func TestLoadGoods(t *testing.T) {
	withTestDB(t, &orm.GuildShopGood{})
	seedGoods := []orm.GuildShopGood{
		{CommanderID: 50, Index: 1, GoodsID: 10, Count: 1},
		{CommanderID: 50, Index: 2, GoodsID: 11, Count: 2},
	}
	if err := orm.GormDB.Create(&seedGoods).Error; err != nil {
		t.Fatalf("seed goods failed: %v", err)
	}

	goods, err := LoadGoods(50)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(goods) != 2 {
		t.Fatalf("expected 2 goods, got %d", len(goods))
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
