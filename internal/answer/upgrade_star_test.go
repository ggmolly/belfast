package answer

import (
	"fmt"
	"os"
	"testing"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func setupUpgradeStarTest(t *testing.T) *connection.Client {
	t.Helper()
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearTable(t, &orm.OwnedShipEquipment{})
	clearTable(t, &orm.OwnedShip{})
	clearTable(t, &orm.OwnedEquipment{})
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.CommanderItem{})
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.Commander{})
	if err := orm.CreateCommanderRoot(701, 701, "Upgrade Star Tester", 0, 0); err != nil {
		t.Fatalf("create commander: %v", err)
	}
	commander := orm.Commander{CommanderID: 701}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func seedBreakoutTemplate(t *testing.T, templateID uint32, groupType uint32, maxLevel uint32) {
	t.Helper()
	payload := fmt.Sprintf(`{"id":%d,"strengthen_id":1,"group_type":%d,"max_level":%d}`, templateID, groupType, maxLevel)
	seedConfigEntry(t, "sharecfgdata/ship_data_template.json", fmt.Sprintf("%d", templateID), payload)
}

func seedBreakoutConfig(t *testing.T, templateID uint32, breakoutID uint32, level uint32, useGold uint32, useChar uint32, useCharNum uint32, useItem string) {
	t.Helper()
	payload := fmt.Sprintf(`{"id":%d,"breakout_id":%d,"pre_id":0,"level":%d,"use_gold":%d,"use_item":%s,"use_char":%d,"use_char_num":%d,"weapon_ids":[]}`, templateID, breakoutID, level, useGold, useItem, useChar, useCharNum)
	seedConfigEntry(t, "sharecfgdata/ship_data_breakout.json", fmt.Sprintf("%d", templateID), payload)
}

func TestUpgradeStarSuccess(t *testing.T) {
	client := setupUpgradeStarTest(t)
	seedBreakoutTemplate(t, 1001, 10, 70)
	seedBreakoutTemplate(t, 1002, 10, 80)
	seedBreakoutTemplate(t, 2001, 10, 70)
	seedBreakoutConfig(t, 1001, 1002, 10, 300, 10, 1, "[[18001,2]]")
	mainShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 15, MaxLevel: 70}
	if err := mainShip.Create(); err != nil {
		t.Fatalf("create main ship: %v", err)
	}
	execAnswerTestSQLT(t, "UPDATE owned_ships SET level = $1, max_level = $2 WHERE owner_id = $3 AND id = $4", int64(mainShip.Level), int64(mainShip.MaxLevel), int64(mainShip.OwnerID), int64(mainShip.ID))
	materialShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 2001, Level: 1}
	if err := materialShip.Create(); err != nil {
		t.Fatalf("create material ship: %v", err)
	}
	execAnswerTestSQLT(t, "UPDATE owned_ships SET level = $1 WHERE owner_id = $2 AND id = $3", int64(materialShip.Level), int64(materialShip.OwnerID), int64(materialShip.ID))
	equipEntry := orm.OwnedShipEquipment{OwnerID: client.Commander.CommanderID, ShipID: materialShip.ID, Pos: 1, EquipID: 3001, SkinID: 0}
	execAnswerTestSQLT(t, "INSERT INTO owned_ship_equipments (owner_id, ship_id, pos, equip_id, skin_id) VALUES ($1, $2, $3, $4, $5)", int64(equipEntry.OwnerID), int64(equipEntry.ShipID), int64(equipEntry.Pos), int64(equipEntry.EquipID), int64(equipEntry.SkinID))
	execAnswerTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1), int64(1000))
	execAnswerTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(18001), int64(2))
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	payload := protobuf.CS_12027{ShipId: proto.Uint32(mainShip.ID), MaterialIdList: []uint32{materialShip.ID}}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := UpgradeStar(&buf, client); err != nil {
		t.Fatalf("UpgradeStar failed: %v", err)
	}
	response := &protobuf.SC_12028{}
	decodePacket(t, client, 12028, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected success result, got %d", response.GetResult())
	}

	updated, err := orm.GetOwnedShipByOwnerAndID(client.Commander.CommanderID, mainShip.ID)
	if err != nil {
		t.Fatalf("load updated ship: %v", err)
	}
	if updated.ShipID != 1002 {
		t.Fatalf("expected ship_id 1002, got %d", updated.ShipID)
	}
	if updated.MaxLevel != 80 {
		t.Fatalf("expected max_level 80, got %d", updated.MaxLevel)
	}
	gold := queryAnswerTestInt64(t, "SELECT amount FROM owned_resources WHERE commander_id = $1 AND resource_id = $2", int64(client.Commander.CommanderID), int64(1))
	if gold != 700 {
		t.Fatalf("expected gold 700, got %d", gold)
	}
	item := queryAnswerTestInt64(t, "SELECT count FROM commander_items WHERE commander_id = $1 AND item_id = $2", int64(client.Commander.CommanderID), int64(18001))
	if item != 0 {
		t.Fatalf("expected item count 0, got %d", item)
	}
	if _, err := orm.GetOwnedShipByOwnerAndID(client.Commander.CommanderID, materialShip.ID); err == nil {
		t.Fatalf("expected material ship to be deleted")
	}
	bag := queryAnswerTestInt64(t, "SELECT count FROM owned_equipments WHERE commander_id = $1 AND equipment_id = $2", int64(client.Commander.CommanderID), int64(3001))
	if bag != 1 {
		t.Fatalf("expected equipment count 1, got %d", bag)
	}
}

func TestUpgradeStarMaterialMismatch(t *testing.T) {
	client := setupUpgradeStarTest(t)
	seedBreakoutTemplate(t, 1001, 10, 70)
	seedBreakoutTemplate(t, 1002, 10, 80)
	seedBreakoutTemplate(t, 2001, 99, 70)
	seedBreakoutConfig(t, 1001, 1002, 10, 300, 10, 1, "[]")
	mainShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 15, MaxLevel: 70}
	if err := mainShip.Create(); err != nil {
		t.Fatalf("create main ship: %v", err)
	}
	execAnswerTestSQLT(t, "UPDATE owned_ships SET level = $1, max_level = $2 WHERE owner_id = $3 AND id = $4", int64(mainShip.Level), int64(mainShip.MaxLevel), int64(mainShip.OwnerID), int64(mainShip.ID))
	materialShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 2001, Level: 1}
	if err := materialShip.Create(); err != nil {
		t.Fatalf("create material ship: %v", err)
	}
	execAnswerTestSQLT(t, "UPDATE owned_ships SET level = $1 WHERE owner_id = $2 AND id = $3", int64(materialShip.Level), int64(materialShip.OwnerID), int64(materialShip.ID))
	execAnswerTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1), int64(1000))
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	payload := protobuf.CS_12027{ShipId: proto.Uint32(mainShip.ID), MaterialIdList: []uint32{materialShip.ID}}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := UpgradeStar(&buf, client); err != nil {
		t.Fatalf("UpgradeStar failed: %v", err)
	}
	response := &protobuf.SC_12028{}
	decodePacket(t, client, 12028, response)
	if response.GetResult() != 1 {
		t.Fatalf("expected failure result, got %d", response.GetResult())
	}
	updated, err := orm.GetOwnedShipByOwnerAndID(client.Commander.CommanderID, mainShip.ID)
	if err != nil {
		t.Fatalf("load updated ship: %v", err)
	}
	if updated.ShipID != 1001 {
		t.Fatalf("expected ship_id 1001, got %d", updated.ShipID)
	}
	if _, err := orm.GetOwnedShipByOwnerAndID(client.Commander.CommanderID, materialShip.ID); err != nil {
		t.Fatalf("expected material ship to remain: %v", err)
	}
}

func TestUpgradeStarDuplicateMaterials(t *testing.T) {
	client := setupUpgradeStarTest(t)
	seedBreakoutTemplate(t, 1001, 10, 70)
	seedBreakoutTemplate(t, 1002, 10, 80)
	seedBreakoutTemplate(t, 2001, 10, 70)
	seedBreakoutConfig(t, 1001, 1002, 10, 0, 10, 2, "[]")
	mainShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 15, MaxLevel: 70}
	if err := mainShip.Create(); err != nil {
		t.Fatalf("create main ship: %v", err)
	}
	execAnswerTestSQLT(t, "UPDATE owned_ships SET level = $1, max_level = $2 WHERE owner_id = $3 AND id = $4", int64(mainShip.Level), int64(mainShip.MaxLevel), int64(mainShip.OwnerID), int64(mainShip.ID))
	materialShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 2001, Level: 1}
	if err := materialShip.Create(); err != nil {
		t.Fatalf("create material ship: %v", err)
	}
	execAnswerTestSQLT(t, "UPDATE owned_ships SET level = $1 WHERE owner_id = $2 AND id = $3", int64(materialShip.Level), int64(materialShip.OwnerID), int64(materialShip.ID))
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	payload := protobuf.CS_12027{ShipId: proto.Uint32(mainShip.ID), MaterialIdList: []uint32{materialShip.ID, materialShip.ID}}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := UpgradeStar(&buf, client); err != nil {
		t.Fatalf("UpgradeStar failed: %v", err)
	}
	response := &protobuf.SC_12028{}
	decodePacket(t, client, 12028, response)
	if response.GetResult() != 1 {
		t.Fatalf("expected failure result, got %d", response.GetResult())
	}
	if _, err := orm.GetOwnedShipByOwnerAndID(client.Commander.CommanderID, materialShip.ID); err != nil {
		t.Fatalf("expected material ship to remain: %v", err)
	}
}
