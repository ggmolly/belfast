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
	execAnswerExternalTestSQLT(t, "DELETE FROM config_entries WHERE category = $1", guildStoreConfigCategory)
	execAnswerExternalTestSQLT(t, "DELETE FROM config_entries WHERE category = $1", guildSetConfigCategory)
	stores := []guildStoreEntry{{ID: 1, Weight: 100, GoodsPurchaseLimit: 2}, {ID: 2, Weight: 100, GoodsPurchaseLimit: 1}, {ID: 3, Weight: 100, GoodsPurchaseLimit: 5}}
	for _, store := range stores {
		payload, err := json.Marshal(store)
		if err != nil {
			t.Fatalf("failed to marshal guild store: %v", err)
		}
		execAnswerExternalTestSQLT(t, "INSERT INTO config_entries (category, key, data) VALUES ($1, $2, $3::jsonb)", guildStoreConfigCategory, fmt.Sprintf("%d", store.ID), string(payload))
	}
	setEntries := []guildSetEntry{{Key: "store_goods_quantity", KeyValue: 2}, {Key: "store_reset_cost", KeyValue: 50}}
	for _, setEntry := range setEntries {
		payload, err := json.Marshal(setEntry)
		if err != nil {
			t.Fatalf("failed to marshal guild set entry: %v", err)
		}
		execAnswerExternalTestSQLT(t, "INSERT INTO config_entries (category, key, data) VALUES ($1, $2, $3::jsonb)", guildSetConfigCategory, setEntry.Key, string(payload))
	}
}

func setupGuildShopCommander(t *testing.T, commanderID uint32) *orm.Commander {
	name := fmt.Sprintf("Guild Shop Commander %d", commanderID)
	if err := orm.CreateCommanderRoot(commanderID, commanderID, name, 0, 0); err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	resource := orm.OwnedResource{CommanderID: commanderID, ResourceID: 8, Amount: 200}
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(resource.CommanderID), int64(resource.ResourceID), int64(resource.Amount))
	commander := orm.Commander{CommanderID: commanderID}
	if err := commander.Load(); err != nil {
		t.Fatalf("failed to load commander: %v", err)
	}
	commander.OwnedResourcesMap = map[uint32]*orm.OwnedResource{resource.ResourceID: &resource}
	commander.CommanderItemsMap = map[uint32]*orm.CommanderItem{}
	return &commander
}

func cleanupGuildShopData(t *testing.T, commanderID uint32) {
	execAnswerExternalTestSQLT(t, "DELETE FROM guild_shop_states WHERE commander_id = $1", int64(commanderID))
	execAnswerExternalTestSQLT(t, "DELETE FROM guild_shop_goods WHERE commander_id = $1", int64(commanderID))
	execAnswerExternalTestSQLT(t, "DELETE FROM owned_resources WHERE commander_id = $1", int64(commanderID))
	execAnswerExternalTestSQLT(t, "DELETE FROM commanders WHERE commander_id = $1", int64(commanderID))
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

	execAnswerExternalTestSQLT(t, "INSERT INTO guild_shop_states (commander_id, refresh_count, next_refresh_time) VALUES ($1, $2, $3)", int64(commanderID), int64(0), int64(time.Now().Add(time.Hour).Unix()))

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
