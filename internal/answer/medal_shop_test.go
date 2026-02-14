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
	execAnswerExternalTestSQLT(t, "DELETE FROM config_entries WHERE category = $1", monthShopConfigCategory)
	execAnswerExternalTestSQLT(t, "DELETE FROM config_entries WHERE category = $1", shopTemplateCategory)
	monthPayload, err := json.Marshal(medalMonthShopTemplate{HonorMedalShopGoods: []uint32{10000, 10001}})
	if err != nil {
		t.Fatalf("failed to marshal month shop template: %v", err)
	}
	execAnswerExternalTestSQLT(t, "INSERT INTO config_entries (category, key, data) VALUES ($1, $2, $3::jsonb)", monthShopConfigCategory, "1", string(monthPayload))
	entries := []medalShopTemplateEntry{{ID: 10000, GoodsPurchaseLimit: 5}, {ID: 10001, GoodsPurchaseLimit: 2}}
	for _, entry := range entries {
		payload, err := json.Marshal(entry)
		if err != nil {
			t.Fatalf("failed to marshal shop template entry: %v", err)
		}
		execAnswerExternalTestSQLT(t, "INSERT INTO config_entries (category, key, data) VALUES ($1, $2, $3::jsonb)", shopTemplateCategory, fmt.Sprintf("%d", entry.ID), string(payload))
	}
}

func setupMedalShopCommander(t *testing.T, commanderID uint32) *orm.Commander {
	name := fmt.Sprintf("Medal Shop Commander %d", commanderID)
	if err := orm.CreateCommanderRoot(commanderID, commanderID, name, 0, 0); err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	commander := orm.Commander{CommanderID: commanderID}
	if err := commander.Load(); err != nil {
		t.Fatalf("failed to load commander: %v", err)
	}
	commander.OwnedResourcesMap = map[uint32]*orm.OwnedResource{}
	commander.CommanderItemsMap = map[uint32]*orm.CommanderItem{}
	return &commander
}

func cleanupMedalShopData(t *testing.T, commanderID uint32) {
	execAnswerExternalTestSQLT(t, "DELETE FROM medal_shop_goods WHERE commander_id = $1", int64(commanderID))
	execAnswerExternalTestSQLT(t, "DELETE FROM medal_shop_states WHERE commander_id = $1", int64(commanderID))
	execAnswerExternalTestSQLT(t, "DELETE FROM commanders WHERE commander_id = $1", int64(commanderID))
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

	execAnswerExternalTestSQLT(t, "INSERT INTO medal_shop_states (commander_id, next_refresh_time) VALUES ($1, $2)", int64(commanderID), int64(time.Now().Add(-time.Hour).Unix()))
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
