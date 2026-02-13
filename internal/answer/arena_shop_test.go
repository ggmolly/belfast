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

const arenaShopConfigCategory = "ShareCfg/arena_data_shop.json"

type arenaShopTemplate struct {
	CommodityList1      [][]uint32 `json:"commodity_list_1"`
	CommodityList2      [][]uint32 `json:"commodity_list_2"`
	CommodityList3      [][]uint32 `json:"commodity_list_3"`
	CommodityList4      [][]uint32 `json:"commodity_list_4"`
	CommodityList5      [][]uint32 `json:"commodity_list_5"`
	CommodityListCommon [][]uint32 `json:"commodity_list_common"`
	RefreshPrice        []uint32   `json:"refresh_price"`
}

func seedArenaShopConfig(t *testing.T, template arenaShopTemplate) {
	data, err := json.Marshal(template)
	if err != nil {
		t.Fatalf("failed to marshal shop config: %v", err)
	}
	execAnswerExternalTestSQLT(t, "DELETE FROM config_entries WHERE category = $1", arenaShopConfigCategory)
	execAnswerExternalTestSQLT(t, "INSERT INTO config_entries (category, key, data) VALUES ($1, $2, $3::jsonb)", arenaShopConfigCategory, "1", string(data))
}

func setupArenaShopCommander(t *testing.T, commanderID uint32) *orm.Commander {
	name := fmt.Sprintf("Arena Shop Commander %d", commanderID)
	if err := orm.CreateCommanderRoot(commanderID, commanderID, name, 0, 0); err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	resource := orm.OwnedResource{
		CommanderID: commanderID,
		ResourceID:  4,
		Amount:      200,
	}
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(resource.CommanderID), int64(resource.ResourceID), int64(resource.Amount))
	commander := orm.Commander{CommanderID: commanderID}
	if err := commander.Load(); err != nil {
		t.Fatalf("failed to load commander: %v", err)
	}
	commander.OwnedResourcesMap = map[uint32]*orm.OwnedResource{resource.ResourceID: &resource}
	commander.CommanderItemsMap = map[uint32]*orm.CommanderItem{}
	return &commander
}

func cleanupArenaShopData(t *testing.T, commanderID uint32) {
	execAnswerExternalTestSQLT(t, "DELETE FROM arena_shop_states WHERE commander_id = $1", int64(commanderID))
	execAnswerExternalTestSQLT(t, "DELETE FROM owned_resources WHERE commander_id = $1", int64(commanderID))
	execAnswerExternalTestSQLT(t, "DELETE FROM commanders WHERE commander_id = $1", int64(commanderID))
}

func TestGetArenaShopCreatesState(t *testing.T) {
	commanderID := uint32(6001)
	cleanupArenaShopData(t, commanderID)
	seedArenaShopConfig(t, arenaShopTemplate{
		CommodityList1:      [][]uint32{{1001, 1}},
		CommodityList2:      [][]uint32{{1002, 1}},
		CommodityList3:      [][]uint32{},
		CommodityList4:      [][]uint32{},
		CommodityList5:      [][]uint32{},
		CommodityListCommon: [][]uint32{{2001, 2}, {2002, 3}},
		RefreshPrice:        []uint32{20, 50},
	})
	client := &connection.Client{Commander: setupArenaShopCommander(t, commanderID)}
	defer cleanupArenaShopData(t, commanderID)

	payload := &protobuf.CS_18100{Type: proto.Uint32(0)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.GetArenaShop(&buf, client); err != nil {
		t.Fatalf("GetArenaShop failed: %v", err)
	}
	response := &protobuf.SC_18101{}
	decodeTestPacket(t, client, 18101, response)
	if response.GetFlashCount() != 0 {
		t.Fatalf("expected flash_count 0, got %d", response.GetFlashCount())
	}
	if len(response.GetArenaShopList()) != 3 {
		t.Fatalf("expected 3 shop entries, got %d", len(response.GetArenaShopList()))
	}
	if response.GetNextFlashTime() <= uint32(time.Now().Unix()) {
		t.Fatalf("expected next flash time in the future")
	}
}

func TestRefreshArenaShopConsumesGems(t *testing.T) {
	commanderID := uint32(6002)
	cleanupArenaShopData(t, commanderID)
	seedArenaShopConfig(t, arenaShopTemplate{
		CommodityList1:      [][]uint32{{1001, 1}},
		CommodityList2:      [][]uint32{{1002, 1}},
		CommodityList3:      [][]uint32{},
		CommodityList4:      [][]uint32{},
		CommodityList5:      [][]uint32{},
		CommodityListCommon: [][]uint32{{2001, 2}},
		RefreshPrice:        []uint32{20, 50},
	})
	client := &connection.Client{Commander: setupArenaShopCommander(t, commanderID)}
	defer cleanupArenaShopData(t, commanderID)

	execAnswerExternalTestSQLT(
		t,
		"INSERT INTO arena_shop_states (commander_id, flash_count, last_refresh_time, next_flash_time) VALUES ($1, $2, $3, $4)",
		int64(commanderID),
		int64(0),
		int64(time.Now().Unix()),
		int64(time.Now().Add(time.Hour).Unix()),
	)

	payload := &protobuf.CS_18102{Type: proto.Uint32(0)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.RefreshArenaShop(&buf, client); err != nil {
		t.Fatalf("RefreshArenaShop failed: %v", err)
	}
	response := &protobuf.SC_18103{}
	decodeTestPacket(t, client, 18103, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if len(response.GetArenaShopList()) != 2 {
		t.Fatalf("expected 2 shop entries, got %d", len(response.GetArenaShopList()))
	}
	if client.Commander.GetResourceCount(4) != 180 {
		t.Fatalf("expected gem count 180, got %d", client.Commander.GetResourceCount(4))
	}
}

func TestGetArenaShopResetsOnExpiry(t *testing.T) {
	commanderID := uint32(6003)
	cleanupArenaShopData(t, commanderID)
	seedArenaShopConfig(t, arenaShopTemplate{
		CommodityList1:      [][]uint32{{1001, 1}},
		CommodityList2:      [][]uint32{{1002, 1}},
		CommodityList3:      [][]uint32{},
		CommodityList4:      [][]uint32{},
		CommodityList5:      [][]uint32{},
		CommodityListCommon: [][]uint32{{2001, 2}},
		RefreshPrice:        []uint32{20, 50},
	})
	client := &connection.Client{Commander: setupArenaShopCommander(t, commanderID)}
	defer cleanupArenaShopData(t, commanderID)

	execAnswerExternalTestSQLT(
		t,
		"INSERT INTO arena_shop_states (commander_id, flash_count, last_refresh_time, next_flash_time) VALUES ($1, $2, $3, $4)",
		int64(commanderID),
		int64(3),
		int64(time.Now().Add(-24*time.Hour).Unix()),
		int64(time.Now().Add(-time.Hour).Unix()),
	)

	payload := &protobuf.CS_18100{Type: proto.Uint32(0)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.GetArenaShop(&buf, client); err != nil {
		t.Fatalf("GetArenaShop failed: %v", err)
	}
	response := &protobuf.SC_18101{}
	decodeTestPacket(t, client, 18101, response)
	if response.GetFlashCount() != 0 {
		t.Fatalf("expected flash_count 0 after reset, got %d", response.GetFlashCount())
	}
}
