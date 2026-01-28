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

const miniGameShopCategory = "ShareCfg/gameroom_shop_template.json"

type miniGameShopEntry struct {
	ID                 uint32     `json:"id"`
	GoodsPurchaseLimit uint32     `json:"goods_purchase_limit"`
	Order              uint32     `json:"order"`
	Time               [][][3]int `json:"time"`
}

func seedMiniGameShopConfig(t *testing.T, entries []miniGameShopEntry) {
	orm.GormDB.Where("category = ?", miniGameShopCategory).Delete(&orm.ConfigEntry{})
	for _, entry := range entries {
		payload, err := json.Marshal(entry)
		if err != nil {
			t.Fatalf("failed to marshal minigame shop entry: %v", err)
		}
		config := orm.ConfigEntry{Category: miniGameShopCategory, Key: fmt.Sprintf("%d", entry.ID), Data: payload}
		if err := orm.GormDB.Create(&config).Error; err != nil {
			t.Fatalf("failed to create minigame shop entry: %v", err)
		}
	}
}

func setupMiniGameCommander(t *testing.T, commanderID uint32) *orm.Commander {
	commander := orm.Commander{CommanderID: commanderID, AccountID: commanderID, Name: fmt.Sprintf("MiniGame Commander %d", commanderID)}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	commander.OwnedResourcesMap = map[uint32]*orm.OwnedResource{}
	commander.CommanderItemsMap = map[uint32]*orm.CommanderItem{}
	return &commander
}

func cleanupMiniGameShopData(t *testing.T, commanderID uint32) {
	if err := orm.GormDB.Where("commander_id = ?", commanderID).Delete(&orm.MiniGameShopGood{}).Error; err != nil {
		t.Fatalf("failed to cleanup minigame shop goods: %v", err)
	}
	if err := orm.GormDB.Where("commander_id = ?", commanderID).Delete(&orm.MiniGameShopState{}).Error; err != nil {
		t.Fatalf("failed to cleanup minigame shop state: %v", err)
	}
	if err := orm.GormDB.Unscoped().Delete(&orm.Commander{}, commanderID).Error; err != nil {
		t.Fatalf("failed to cleanup commander: %v", err)
	}
}

func TestGetMiniGameShopFiltersByTime(t *testing.T) {
	commanderID := uint32(9001)
	cleanupMiniGameShopData(t, commanderID)
	now := time.Now().UTC()
	seedMiniGameShopConfig(t, []miniGameShopEntry{
		{
			ID:                 1,
			GoodsPurchaseLimit: 2,
			Order:              1,
			Time:               [][][3]int{{{now.Year() - 1, int(now.Month()), now.Day()}, {now.Year() + 1, int(now.Month()), now.Day()}}},
		},
		{
			ID:                 2,
			GoodsPurchaseLimit: 3,
			Order:              2,
			Time:               [][][3]int{{{now.Year() - 2, int(now.Month()), now.Day()}, {now.Year() - 1, int(now.Month()), now.Day()}}},
		},
	})
	client := &connection.Client{Commander: setupMiniGameCommander(t, commanderID)}
	defer cleanupMiniGameShopData(t, commanderID)

	payload := &protobuf.CS_26150{Type: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.GetMiniGameShop(&buf, client); err != nil {
		t.Fatalf("GetMiniGameShop failed: %v", err)
	}
	response := &protobuf.SC_26151{}
	decodeTestPacket(t, client, 26151, response)
	if len(response.GetGoods()) != 1 {
		t.Fatalf("expected 1 good, got %d", len(response.GetGoods()))
	}
	if response.GetGoods()[0].GetCount() != 2 {
		t.Fatalf("expected count 2, got %d", response.GetGoods()[0].GetCount())
	}
}

func TestGetMiniGameShopResetsOnExpiry(t *testing.T) {
	commanderID := uint32(9002)
	cleanupMiniGameShopData(t, commanderID)
	now := time.Now().UTC()
	seedMiniGameShopConfig(t, []miniGameShopEntry{
		{
			ID:                 3,
			GoodsPurchaseLimit: 1,
			Order:              1,
			Time:               [][][3]int{{{now.Year() - 1, int(now.Month()), now.Day()}, {now.Year() + 1, int(now.Month()), now.Day()}}},
		},
	})
	client := &connection.Client{Commander: setupMiniGameCommander(t, commanderID)}
	defer cleanupMiniGameShopData(t, commanderID)

	state := orm.MiniGameShopState{CommanderID: commanderID, NextRefreshTime: uint32(time.Now().Add(-time.Hour).Unix())}
	if err := orm.GormDB.Create(&state).Error; err != nil {
		t.Fatalf("failed to create minigame shop state: %v", err)
	}
	payload := &protobuf.CS_26150{Type: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.GetMiniGameShop(&buf, client); err != nil {
		t.Fatalf("GetMiniGameShop failed: %v", err)
	}
	response := &protobuf.SC_26151{}
	decodeTestPacket(t, client, 26151, response)
	if response.GetNextFlashTime() <= uint32(time.Now().Unix()) {
		t.Fatalf("expected next_flash_time refreshed")
	}
}
