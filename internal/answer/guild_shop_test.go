package answer_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

const (
	guildStoreConfigCategory = "ShareCfg/guild_store.json"
	guildSetConfigCategory   = "ShareCfg/guildset.json"
)

type guildStoreEntry struct {
	ID                 uint32 `json:"id"`
	Weight             uint32 `json:"weight"`
	GoodsPurchaseLimit uint32 `json:"goods_purchase_limit"`
}

type guildSetEntry struct {
	Key      string `json:"key"`
	KeyValue uint32 `json:"key_value"`
}

func seedGuildShopConfig(t *testing.T) {
	orm.GormDB.Where("category = ?", guildStoreConfigCategory).Delete(&orm.ConfigEntry{})
	orm.GormDB.Where("category = ?", guildSetConfigCategory).Delete(&orm.ConfigEntry{})
	stores := []guildStoreEntry{{ID: 1, Weight: 100, GoodsPurchaseLimit: 2}, {ID: 2, Weight: 100, GoodsPurchaseLimit: 1}, {ID: 3, Weight: 100, GoodsPurchaseLimit: 5}}
	for _, store := range stores {
		payload, err := json.Marshal(store)
		if err != nil {
			t.Fatalf("failed to marshal guild store: %v", err)
		}
		entry := orm.ConfigEntry{Category: guildStoreConfigCategory, Key: fmt.Sprintf("%d", store.ID), Data: payload}
		if err := orm.GormDB.Create(&entry).Error; err != nil {
			t.Fatalf("failed to create guild store entry: %v", err)
		}
	}
	setEntries := []guildSetEntry{{Key: "store_goods_quantity", KeyValue: 2}, {Key: "store_reset_cost", KeyValue: 50}}
	for _, setEntry := range setEntries {
		payload, err := json.Marshal(setEntry)
		if err != nil {
			t.Fatalf("failed to marshal guild set entry: %v", err)
		}
		entry := orm.ConfigEntry{Category: guildSetConfigCategory, Key: setEntry.Key, Data: payload}
		if err := orm.GormDB.Create(&entry).Error; err != nil {
			t.Fatalf("failed to create guild set entry: %v", err)
		}
	}
}

func setupGuildShopCommander(t *testing.T, commanderID uint32) *orm.Commander {
	commander := orm.Commander{CommanderID: commanderID, AccountID: commanderID, Name: fmt.Sprintf("Guild Shop Commander %d", commanderID)}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	resource := orm.OwnedResource{CommanderID: commanderID, ResourceID: 8, Amount: 200}
	if err := orm.GormDB.Create(&resource).Error; err != nil {
		t.Fatalf("failed to create guild coin resource: %v", err)
	}
	commander.OwnedResourcesMap = map[uint32]*orm.OwnedResource{resource.ResourceID: &resource}
	commander.CommanderItemsMap = map[uint32]*orm.CommanderItem{}
	return &commander
}

func cleanupGuildShopData(t *testing.T, commanderID uint32) {
	if err := orm.GormDB.Where("commander_id = ?", commanderID).Delete(&orm.GuildShopState{}).Error; err != nil {
		t.Fatalf("failed to cleanup guild shop state: %v", err)
	}
	if err := orm.GormDB.Where("commander_id = ?", commanderID).Delete(&orm.GuildShopGood{}).Error; err != nil {
		t.Fatalf("failed to cleanup guild shop goods: %v", err)
	}
	if err := orm.GormDB.Where("commander_id = ?", commanderID).Delete(&orm.OwnedResource{}).Error; err != nil {
		t.Fatalf("failed to cleanup resources: %v", err)
	}
	if err := orm.GormDB.Unscoped().Delete(&orm.Commander{}, commanderID).Error; err != nil {
		t.Fatalf("failed to cleanup commander: %v", err)
	}
}

func TestGetGuildShopCreatesState(t *testing.T) {
	commanderID := uint32(7001)
	cleanupGuildShopData(t, commanderID)
	seedGuildShopConfig(t)
	client := &connection.Client{Commander: setupGuildShopCommander(t, commanderID)}
	defer cleanupGuildShopData(t, commanderID)

	payload := &protobuf.CS_60033{Type: proto.Uint32(0)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.GetGuildShop(&buf, client); err != nil {
		t.Fatalf("GetGuildShop failed: %v", err)
	}
	response := &protobuf.SC_60034{}
	decodeTestPacket(t, client, 60034, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if response.GetInfo().GetRefreshCount() != 0 {
		t.Fatalf("expected refresh_count 0, got %d", response.GetInfo().GetRefreshCount())
	}
	if len(response.GetInfo().GetGoodList()) != 2 {
		t.Fatalf("expected 2 goods, got %d", len(response.GetInfo().GetGoodList()))
	}
	if response.GetInfo().GetNextRefreshTime() <= uint32(time.Now().Unix()) {
		t.Fatalf("expected next refresh time in the future")
	}
}

func TestGuildShopManualRefreshConsumesCoins(t *testing.T) {
	commanderID := uint32(7002)
	cleanupGuildShopData(t, commanderID)
	seedGuildShopConfig(t)
	client := &connection.Client{Commander: setupGuildShopCommander(t, commanderID)}
	defer cleanupGuildShopData(t, commanderID)

	state := orm.GuildShopState{CommanderID: commanderID, RefreshCount: 0, NextRefreshTime: uint32(time.Now().Add(time.Hour).Unix())}
	if err := orm.GormDB.Create(&state).Error; err != nil {
		t.Fatalf("failed to create guild shop state: %v", err)
	}

	payload := &protobuf.CS_60033{Type: proto.Uint32(2)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.GetGuildShop(&buf, client); err != nil {
		t.Fatalf("GetGuildShop failed: %v", err)
	}
	response := &protobuf.SC_60034{}
	decodeTestPacket(t, client, 60034, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if response.GetInfo().GetRefreshCount() != 1 {
		t.Fatalf("expected refresh_count 1, got %d", response.GetInfo().GetRefreshCount())
	}
	if client.Commander.GetResourceCount(8) != 150 {
		t.Fatalf("expected guild coin count 150, got %d", client.Commander.GetResourceCount(8))
	}
}
