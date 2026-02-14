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
	execAnswerExternalTestSQLT(t, "DELETE FROM config_entries WHERE category = $1", miniGameShopCategory)
	for _, entry := range entries {
		payload, err := json.Marshal(entry)
		if err != nil {
			t.Fatalf("failed to marshal minigame shop entry: %v", err)
		}
		execAnswerExternalTestSQLT(t, "INSERT INTO config_entries (category, key, data) VALUES ($1, $2, $3::jsonb)", miniGameShopCategory, fmt.Sprintf("%d", entry.ID), string(payload))
	}
}

func setupMiniGameCommander(t *testing.T, commanderID uint32) *orm.Commander {
	name := fmt.Sprintf("MiniGame Commander %d", commanderID)
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

func cleanupMiniGameShopData(t *testing.T, commanderID uint32) {
	execAnswerExternalTestSQLT(t, "DELETE FROM mini_game_shop_goods WHERE commander_id = $1", int64(commanderID))
	execAnswerExternalTestSQLT(t, "DELETE FROM mini_game_shop_states WHERE commander_id = $1", int64(commanderID))
	execAnswerExternalTestSQLT(t, "DELETE FROM commanders WHERE commander_id = $1", int64(commanderID))
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

	execAnswerExternalTestSQLT(t, "INSERT INTO mini_game_shop_states (commander_id, next_refresh_time) VALUES ($1, $2)", int64(commanderID), int64(time.Now().Add(-time.Hour).Unix()))
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
