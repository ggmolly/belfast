package answer

import (
	"fmt"
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func seedShipLevelEntry(t *testing.T, level uint32, levelLimit uint32, rarity uint32, needItems string) {
	t.Helper()
	key := fmt.Sprintf("%d", level)
	if needItems != "" {
		payload := fmt.Sprintf(`{"level":%d,"exp":10,"exp_ur":10,"level_limit":%d,"need_item_rarity%d":%s}`, level, levelLimit, rarity, needItems)
		seedConfigEntry(t, "ShareCfg/ship_level.json", key, payload)
		return
	}
	payload := fmt.Sprintf(`{"level":%d,"exp":10,"exp_ur":10,"level_limit":%d}`, level, levelLimit)
	seedConfigEntry(t, "ShareCfg/ship_level.json", key, payload)
}

func TestUpgradeShipMaxLevelSuccessPersistsAndConsumes(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.CommanderItem{})
	clearTable(t, &orm.Ship{})

	seedShipLevelEntry(t, 100, 1, 3, `[[1,1,300],[2,18001,2]]`)
	seedShipLevelEntry(t, 101, 0, 3, "")
	seedShipLevelEntry(t, 102, 0, 3, "")
	seedShipLevelEntry(t, 103, 0, 3, "")
	seedShipLevelEntry(t, 104, 0, 3, "")
	seedShipLevelEntry(t, 105, 1, 3, "")

	execAnswerTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(1001), "Test Ship", "Test Ship", int64(3), int64(1), int64(1), int64(1), int64(1))
	owned := orm.OwnedShip{ID: 101, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 100, MaxLevel: 100, Exp: 0, SurplusExp: 0, Energy: 150}
	execAnswerTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, level, max_level, exp, surplus_exp, energy, create_time, change_name_timestamp) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())", int64(owned.ID), int64(owned.OwnerID), int64(owned.ShipID), int64(owned.Level), int64(owned.MaxLevel), int64(owned.Exp), int64(owned.SurplusExp), int64(owned.Energy))
	execAnswerTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1), int64(1000))
	execAnswerTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(18001), int64(2))
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	payload := protobuf.CS_12038{ShipId: proto.Uint32(owned.ID)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := UpgradeShipMaxLevel(&buf, client); err != nil {
		t.Fatalf("UpgradeShipMaxLevel failed: %v", err)
	}
	response := &protobuf.SC_12039{}
	decodePacket(t, client, 12039, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected success result, got %d", response.GetResult())
	}

	updated, err := orm.GetOwnedShipByOwnerAndID(client.Commander.CommanderID, owned.ID)
	if err != nil {
		t.Fatalf("load updated ship: %v", err)
	}
	if updated.MaxLevel != 105 {
		t.Fatalf("expected max_level 105, got %d", updated.MaxLevel)
	}
	gold := queryAnswerTestInt64(t, "SELECT amount FROM owned_resources WHERE commander_id = $1 AND resource_id = $2", int64(client.Commander.CommanderID), int64(1))
	if gold != 700 {
		t.Fatalf("expected gold 700, got %d", gold)
	}
	item := queryAnswerTestInt64(t, "SELECT count FROM commander_items WHERE commander_id = $1 AND item_id = $2", int64(client.Commander.CommanderID), int64(18001))
	if item != 0 {
		t.Fatalf("expected item count 0, got %d", item)
	}
}

func TestUpgradeShipMaxLevelInsufficientResourcesNoMutation(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.CommanderItem{})
	clearTable(t, &orm.Ship{})

	seedShipLevelEntry(t, 100, 1, 3, `[[1,1,300],[2,18001,2]]`)
	seedShipLevelEntry(t, 101, 0, 3, "")
	seedShipLevelEntry(t, 102, 0, 3, "")
	seedShipLevelEntry(t, 103, 0, 3, "")
	seedShipLevelEntry(t, 104, 0, 3, "")
	seedShipLevelEntry(t, 105, 1, 3, "")

	execAnswerTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(1001), "Test Ship", "Test Ship", int64(3), int64(1), int64(1), int64(1), int64(1))
	owned := orm.OwnedShip{ID: 101, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 100, MaxLevel: 100, Exp: 0, SurplusExp: 0, Energy: 150}
	execAnswerTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, level, max_level, exp, surplus_exp, energy, create_time, change_name_timestamp) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())", int64(owned.ID), int64(owned.OwnerID), int64(owned.ShipID), int64(owned.Level), int64(owned.MaxLevel), int64(owned.Exp), int64(owned.SurplusExp), int64(owned.Energy))
	execAnswerTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1), int64(100))
	execAnswerTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(18001), int64(2))
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	payload := protobuf.CS_12038{ShipId: proto.Uint32(owned.ID)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := UpgradeShipMaxLevel(&buf, client); err != nil {
		t.Fatalf("UpgradeShipMaxLevel failed: %v", err)
	}
	response := &protobuf.SC_12039{}
	decodePacket(t, client, 12039, response)
	if response.GetResult() == 0 {
		t.Fatalf("expected failure result")
	}

	updated, err := orm.GetOwnedShipByOwnerAndID(client.Commander.CommanderID, owned.ID)
	if err != nil {
		t.Fatalf("load updated ship: %v", err)
	}
	if updated.MaxLevel != 100 {
		t.Fatalf("expected max_level 100, got %d", updated.MaxLevel)
	}
}

func TestUpgradeShipMaxLevelInvalidState(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.CommanderItem{})
	clearTable(t, &orm.Ship{})

	seedShipLevelEntry(t, 100, 1, 3, `[[1,1,300],[2,18001,2]]`)
	seedShipLevelEntry(t, 101, 0, 3, "")
	seedShipLevelEntry(t, 102, 0, 3, "")
	seedShipLevelEntry(t, 103, 0, 3, "")
	seedShipLevelEntry(t, 104, 0, 3, "")
	seedShipLevelEntry(t, 105, 1, 3, "")

	execAnswerTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(1001), "Test Ship", "Test Ship", int64(3), int64(1), int64(1), int64(1), int64(1))
	owned := orm.OwnedShip{ID: 101, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 99, MaxLevel: 100, Exp: 0, SurplusExp: 0, Energy: 150}
	execAnswerTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, level, max_level, exp, surplus_exp, energy, create_time, change_name_timestamp) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())", int64(owned.ID), int64(owned.OwnerID), int64(owned.ShipID), int64(owned.Level), int64(owned.MaxLevel), int64(owned.Exp), int64(owned.SurplusExp), int64(owned.Energy))
	execAnswerTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1), int64(1000))
	execAnswerTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(18001), int64(2))
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	payload := protobuf.CS_12038{ShipId: proto.Uint32(owned.ID)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := UpgradeShipMaxLevel(&buf, client); err != nil {
		t.Fatalf("UpgradeShipMaxLevel failed: %v", err)
	}
	response := &protobuf.SC_12039{}
	decodePacket(t, client, 12039, response)
	if response.GetResult() == 0 {
		t.Fatalf("expected failure result")
	}
}

func TestUpgradeShipMaxLevelAppliesOverflowExp(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.CommanderItem{})
	clearTable(t, &orm.Ship{})

	seedShipLevelEntry(t, 100, 1, 3, `[[1,1,300],[2,18001,2]]`)
	seedShipLevelEntry(t, 101, 0, 3, "")
	seedShipLevelEntry(t, 102, 0, 3, "")
	seedShipLevelEntry(t, 103, 0, 3, "")
	seedShipLevelEntry(t, 104, 0, 3, "")
	seedShipLevelEntry(t, 105, 1, 3, "")

	execAnswerTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(1001), "Test Ship", "Test Ship", int64(3), int64(1), int64(1), int64(1), int64(1))
	owned := orm.OwnedShip{ID: 101, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 100, MaxLevel: 100, Exp: 0, SurplusExp: 35, Energy: 150}
	execAnswerTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, level, max_level, exp, surplus_exp, energy, create_time, change_name_timestamp) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())", int64(owned.ID), int64(owned.OwnerID), int64(owned.ShipID), int64(owned.Level), int64(owned.MaxLevel), int64(owned.Exp), int64(owned.SurplusExp), int64(owned.Energy))
	execAnswerTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1), int64(1000))
	execAnswerTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(18001), int64(2))
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	payload := protobuf.CS_12038{ShipId: proto.Uint32(owned.ID)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := UpgradeShipMaxLevel(&buf, client); err != nil {
		t.Fatalf("UpgradeShipMaxLevel failed: %v", err)
	}
	response := &protobuf.SC_12039{}
	decodePacket(t, client, 12039, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected success result, got %d", response.GetResult())
	}

	updated, err := orm.GetOwnedShipByOwnerAndID(client.Commander.CommanderID, owned.ID)
	if err != nil {
		t.Fatalf("load updated ship: %v", err)
	}
	if updated.MaxLevel != 105 {
		t.Fatalf("expected max_level 105, got %d", updated.MaxLevel)
	}
	if updated.Level != 103 {
		t.Fatalf("expected level 103, got %d", updated.Level)
	}
	if updated.Exp != 5 {
		t.Fatalf("expected exp 5, got %d", updated.Exp)
	}
	if updated.SurplusExp != 0 {
		t.Fatalf("expected surplus exp 0, got %d", updated.SurplusExp)
	}
}
