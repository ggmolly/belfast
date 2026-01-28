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
	monthShopConfigCategory = "ShareCfg/month_shop_template.json"
	shopTemplateCategory    = "ShareCfg/shop_template.json"
)

type medalMonthShopTemplate struct {
	HonorMedalShopGoods []uint32 `json:"honormedal_shop_goods"`
}

type medalShopTemplateEntry struct {
	ID                 uint32 `json:"id"`
	GoodsPurchaseLimit uint32 `json:"goods_purchase_limit"`
}

func seedMedalShopConfig(t *testing.T) {
	orm.GormDB.Where("category = ?", monthShopConfigCategory).Delete(&orm.ConfigEntry{})
	orm.GormDB.Where("category = ?", shopTemplateCategory).Delete(&orm.ConfigEntry{})
	monthPayload, err := json.Marshal(medalMonthShopTemplate{HonorMedalShopGoods: []uint32{10000, 10001}})
	if err != nil {
		t.Fatalf("failed to marshal month shop template: %v", err)
	}
	if err := orm.GormDB.Create(&orm.ConfigEntry{Category: monthShopConfigCategory, Key: "1", Data: monthPayload}).Error; err != nil {
		t.Fatalf("failed to create month shop entry: %v", err)
	}
	entries := []medalShopTemplateEntry{{ID: 10000, GoodsPurchaseLimit: 5}, {ID: 10001, GoodsPurchaseLimit: 2}}
	for _, entry := range entries {
		payload, err := json.Marshal(entry)
		if err != nil {
			t.Fatalf("failed to marshal shop template entry: %v", err)
		}
		if err := orm.GormDB.Create(&orm.ConfigEntry{Category: shopTemplateCategory, Key: fmt.Sprintf("%d", entry.ID), Data: payload}).Error; err != nil {
			t.Fatalf("failed to create shop template entry: %v", err)
		}
	}
}

func setupMedalShopCommander(t *testing.T, commanderID uint32) *orm.Commander {
	commander := orm.Commander{CommanderID: commanderID, AccountID: commanderID, Name: fmt.Sprintf("Medal Shop Commander %d", commanderID)}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	commander.OwnedResourcesMap = map[uint32]*orm.OwnedResource{}
	commander.CommanderItemsMap = map[uint32]*orm.CommanderItem{}
	return &commander
}

func cleanupMedalShopData(t *testing.T, commanderID uint32) {
	if err := orm.GormDB.Where("commander_id = ?", commanderID).Delete(&orm.MedalShopGood{}).Error; err != nil {
		t.Fatalf("failed to cleanup medal shop goods: %v", err)
	}
	if err := orm.GormDB.Where("commander_id = ?", commanderID).Delete(&orm.MedalShopState{}).Error; err != nil {
		t.Fatalf("failed to cleanup medal shop state: %v", err)
	}
	if err := orm.GormDB.Unscoped().Delete(&orm.Commander{}, commanderID).Error; err != nil {
		t.Fatalf("failed to cleanup commander: %v", err)
	}
}

func TestGetMedalShopCreatesState(t *testing.T) {
	commanderID := uint32(8001)
	cleanupMedalShopData(t, commanderID)
	seedMedalShopConfig(t)
	client := &connection.Client{Commander: setupMedalShopCommander(t, commanderID)}
	defer cleanupMedalShopData(t, commanderID)

	payload := &protobuf.CS_16106{Type: proto.Uint32(0)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.GetMedalShop(&buf, client); err != nil {
		t.Fatalf("GetMedalShop failed: %v", err)
	}
	response := &protobuf.SC_16107{}
	decodeTestPacket(t, client, 16107, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if len(response.GetGoodList()) != 2 {
		t.Fatalf("expected 2 goods, got %d", len(response.GetGoodList()))
	}
	if response.GetItemFlashTime() <= uint32(time.Now().Unix()) {
		t.Fatalf("expected item_flash_time in the future")
	}
	if response.GetGoodList()[0].GetCount() != 5 {
		t.Fatalf("expected count 5, got %d", response.GetGoodList()[0].GetCount())
	}
}

func TestGetMedalShopResetsOnExpiry(t *testing.T) {
	commanderID := uint32(8002)
	cleanupMedalShopData(t, commanderID)
	seedMedalShopConfig(t)
	client := &connection.Client{Commander: setupMedalShopCommander(t, commanderID)}
	defer cleanupMedalShopData(t, commanderID)

	state := orm.MedalShopState{CommanderID: commanderID, NextRefreshTime: uint32(time.Now().Add(-time.Hour).Unix())}
	if err := orm.GormDB.Create(&state).Error; err != nil {
		t.Fatalf("failed to create medal shop state: %v", err)
	}
	payload := &protobuf.CS_16106{Type: proto.Uint32(0)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.GetMedalShop(&buf, client); err != nil {
		t.Fatalf("GetMedalShop failed: %v", err)
	}
	response := &protobuf.SC_16107{}
	decodeTestPacket(t, client, 16107, response)
	if response.GetItemFlashTime() <= uint32(time.Now().Unix()) {
		t.Fatalf("expected item_flash_time refreshed")
	}
}
