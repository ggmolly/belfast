package answer_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func setupTransformEquipmentTest(t *testing.T) *connection.Client {
	t.Helper()
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	execAnswerExternalTestSQLT(t, "DELETE FROM owned_ship_equipments")
	execAnswerExternalTestSQLT(t, "DELETE FROM owned_ships")
	execAnswerExternalTestSQLT(t, "DELETE FROM owned_equipments")
	execAnswerExternalTestSQLT(t, "DELETE FROM commander_items")
	execAnswerExternalTestSQLT(t, "DELETE FROM commander_misc_items")
	execAnswerExternalTestSQLT(t, "DELETE FROM owned_resources")
	execAnswerExternalTestSQLT(t, "DELETE FROM equipments")
	execAnswerExternalTestSQLT(t, "DELETE FROM ships")
	execAnswerExternalTestSQLT(t, "DELETE FROM resources")
	execAnswerExternalTestSQLT(t, "DELETE FROM items")
	execAnswerExternalTestSQLT(t, "DELETE FROM config_entries")
	execAnswerExternalTestSQLT(t, "DELETE FROM commanders")

	if err := orm.CreateCommanderRoot(701, 701, "Transform Tester", 0, 0); err != nil {
		t.Fatalf("create commander: %v", err)
	}
	commander := orm.Commander{CommanderID: 701}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func seedResource(t *testing.T, id uint32) {
	t.Helper()
	resource := orm.Resource{ID: id, Name: fmt.Sprintf("res-%d", id)}
	execAnswerExternalTestSQLT(t, "INSERT INTO resources (id, item_id, name) VALUES ($1, $2, $3)", int64(resource.ID), int64(0), resource.Name)
}

func seedItem(t *testing.T, id uint32) {
	t.Helper()
	item := orm.Item{ID: id, Name: fmt.Sprintf("item-%d", id), Rarity: 1, ShopID: -2, Type: 1, VirtualType: 0}
	execAnswerExternalTestSQLT(t, "INSERT INTO items (id, name, rarity, shop_id, type, virtual_type) VALUES ($1, $2, $3, $4, $5, $6)", int64(item.ID), item.Name, int64(item.Rarity), int64(item.ShopID), int64(item.Type), int64(item.VirtualType))
}

func seedCommanderGold(t *testing.T, commanderID uint32, amount uint32) {
	t.Helper()
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(commanderID), int64(1), int64(amount))
}

func seedCommanderItem(t *testing.T, commanderID uint32, itemID uint32, count uint32) {
	t.Helper()
	execAnswerExternalTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(commanderID), int64(itemID), int64(count))
}

func seedEquipUpgradeData(t *testing.T, upgradeID uint32, upgradeFrom uint32, targetID uint32, coinConsume uint32, materials [][]uint32) {
	t.Helper()
	payload, err := json.Marshal(struct {
		ID          uint32     `json:"id"`
		UpgradeFrom uint32     `json:"upgrade_from"`
		TargetID    uint32     `json:"target_id"`
		CoinConsume uint32     `json:"coin_consume"`
		Materials   [][]uint32 `json:"material_consume"`
	}{
		ID:          upgradeID,
		UpgradeFrom: upgradeFrom,
		TargetID:    targetID,
		CoinConsume: coinConsume,
		Materials:   materials,
	})
	if err != nil {
		t.Fatalf("marshal equip upgrade data: %v", err)
	}
	entry := orm.ConfigEntry{Category: "ShareCfg/equip_upgrade_data.json", Key: fmt.Sprintf("%d", upgradeID), Data: payload}
	execAnswerExternalTestSQLT(t, "INSERT INTO config_entries (category, key, data) VALUES ($1, $2, $3::jsonb)", entry.Category, entry.Key, string(entry.Data))
}

func seedEquipment(t *testing.T, id uint32, equipType uint32, equipLimit int) {
	t.Helper()
	equip := orm.Equipment{
		ID:                id,
		DestroyGold:       0,
		DestroyItem:       json.RawMessage(`[]`),
		EquipLimit:        equipLimit,
		Group:             1,
		Important:         0,
		Level:             1,
		Next:              0,
		Prev:              0,
		RestoreGold:       0,
		RestoreItem:       json.RawMessage(`[]`),
		ShipTypeForbidden: json.RawMessage(`[]`),
		TransUseGold:      0,
		TransUseItem:      json.RawMessage(`[]`),
		Type:              equipType,
		UpgradeFormulaID:  json.RawMessage(`[]`),
	}
	execAnswerExternalTestSQLT(t, "INSERT INTO equipments (id, destroy_gold, destroy_item, equip_limit, "+
		"\"group\", important, level, next, prev, restore_gold, restore_item, ship_type_forbidden, trans_use_gold, trans_use_item, type, upgrade_formula_id) VALUES ($1, $2, $3::jsonb, $4, $5, $6, $7, $8, $9, $10, $11::jsonb, $12::jsonb, $13, $14::jsonb, $15, $16::jsonb)",
		int64(equip.ID), int64(equip.DestroyGold), string(equip.DestroyItem), int64(equip.EquipLimit), int64(equip.Group), int64(equip.Important), int64(equip.Level), int64(equip.Next), int64(equip.Prev), int64(equip.RestoreGold), string(equip.RestoreItem), string(equip.ShipTypeForbidden), int64(equip.TransUseGold), string(equip.TransUseItem), int64(equip.Type), string(equip.UpgradeFormulaID))
}

func sendCS14013(t *testing.T, client *connection.Client, shipID uint32, pos uint32, upgradeID uint32) *protobuf.SC_14014 {
	t.Helper()
	payload := protobuf.CS_14013{
		ShipId:    proto.Uint32(shipID),
		Pos:       proto.Uint32(pos),
		UpgradeId: proto.Uint32(upgradeID),
	}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.TransformEquipmentOnShip14013(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	response := &protobuf.SC_14014{}
	decodePacket(t, client, 14014, response)
	return response
}

func TestTransformEquipmentOnShipSuccessAllowed(t *testing.T) {
	client := setupTransformEquipmentTest(t)
	seedResource(t, 1)
	seedItem(t, 3001)
	seedCommanderGold(t, client.Commander.CommanderID, 200)
	seedCommanderItem(t, client.Commander.CommanderID, 3001, 2)

	ship := orm.Ship{TemplateID: 1001, Name: "Ship", EnglishName: "Ship", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	execAnswerExternalTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(ship.TemplateID), ship.Name, ship.EnglishName, int64(ship.RarityID), int64(ship.Star), int64(ship.Type), int64(ship.Nationality), int64(ship.BuildTime))
	seedShipEquipConfig(t, 1001, `{"id":1001,"equip_1":[1],"equip_2":[],"equip_3":[],"equip_4":[],"equip_5":[],"equip_id_1":2001,"equip_id_2":0,"equip_id_3":0}`)
	seedEquipment(t, 2001, 1, 0)
	seedEquipment(t, 2002, 1, 0)
	seedEquipUpgradeData(t, 9001, 2001, 2002, 100, [][]uint32{{3001, 2}})

	ownedShip, err := client.Commander.AddShip(1001)
	if err != nil {
		t.Fatalf("add ship: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	resp := sendCS14013(t, client, ownedShip.ID, 1, 9001)
	if resp.GetResult() != 0 {
		t.Fatalf("expected success result")
	}
	entry, err := orm.GetOwnedShipEquipment(client.Commander.CommanderID, ownedShip.ID, 1)
	if err != nil {
		t.Fatalf("load ship equipment: %v", err)
	}
	if entry.EquipID != 2002 {
		t.Fatalf("expected equip id 2002, got %d", entry.EquipID)
	}
	if client.Commander.GetResourceCount(1) != 100 {
		t.Fatalf("expected gold 100, got %d", client.Commander.GetResourceCount(1))
	}
	if client.Commander.GetItemCount(3001) != 0 {
		t.Fatalf("expected item count 0, got %d", client.Commander.GetItemCount(3001))
	}
}

func TestTransformEquipmentOnShipFailsWrongUpgradePath(t *testing.T) {
	client := setupTransformEquipmentTest(t)
	seedResource(t, 1)
	seedItem(t, 3001)
	seedCommanderGold(t, client.Commander.CommanderID, 200)
	seedCommanderItem(t, client.Commander.CommanderID, 3001, 2)

	ship := orm.Ship{TemplateID: 1001, Name: "Ship", EnglishName: "Ship", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	execAnswerExternalTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(ship.TemplateID), ship.Name, ship.EnglishName, int64(ship.RarityID), int64(ship.Star), int64(ship.Type), int64(ship.Nationality), int64(ship.BuildTime))
	seedShipEquipConfig(t, 1001, `{"id":1001,"equip_1":[1],"equip_2":[],"equip_3":[],"equip_4":[],"equip_5":[],"equip_id_1":2003,"equip_id_2":0,"equip_id_3":0}`)
	seedEquipment(t, 2003, 1, 0)
	seedEquipment(t, 2002, 1, 0)
	seedEquipUpgradeData(t, 9001, 2001, 2002, 100, [][]uint32{{3001, 2}})

	ownedShip, err := client.Commander.AddShip(1001)
	if err != nil {
		t.Fatalf("add ship: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}
	goldBefore := client.Commander.GetResourceCount(1)
	itemBefore := client.Commander.GetItemCount(3001)

	resp := sendCS14013(t, client, ownedShip.ID, 1, 9001)
	if resp.GetResult() == 0 {
		t.Fatalf("expected failure")
	}
	entry, err := orm.GetOwnedShipEquipment(client.Commander.CommanderID, ownedShip.ID, 1)
	if err != nil {
		t.Fatalf("load ship equipment: %v", err)
	}
	if entry.EquipID != 2003 {
		t.Fatalf("expected equip id 2003, got %d", entry.EquipID)
	}
	if client.Commander.GetResourceCount(1) != goldBefore {
		t.Fatalf("expected gold unchanged")
	}
	if client.Commander.GetItemCount(3001) != itemBefore {
		t.Fatalf("expected items unchanged")
	}
}

func TestTransformEquipmentOnShipFailsInsufficientGold(t *testing.T) {
	client := setupTransformEquipmentTest(t)
	seedResource(t, 1)
	seedItem(t, 3001)
	seedCommanderGold(t, client.Commander.CommanderID, 50)
	seedCommanderItem(t, client.Commander.CommanderID, 3001, 2)

	ship := orm.Ship{TemplateID: 1001, Name: "Ship", EnglishName: "Ship", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	execAnswerExternalTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(ship.TemplateID), ship.Name, ship.EnglishName, int64(ship.RarityID), int64(ship.Star), int64(ship.Type), int64(ship.Nationality), int64(ship.BuildTime))
	seedShipEquipConfig(t, 1001, `{"id":1001,"equip_1":[1],"equip_2":[],"equip_3":[],"equip_4":[],"equip_5":[],"equip_id_1":2001,"equip_id_2":0,"equip_id_3":0}`)
	seedEquipment(t, 2001, 1, 0)
	seedEquipment(t, 2002, 1, 0)
	seedEquipUpgradeData(t, 9001, 2001, 2002, 100, [][]uint32{{3001, 2}})

	ownedShip, err := client.Commander.AddShip(1001)
	if err != nil {
		t.Fatalf("add ship: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	resp := sendCS14013(t, client, ownedShip.ID, 1, 9001)
	if resp.GetResult() == 0 {
		t.Fatalf("expected failure")
	}
	entry, err := orm.GetOwnedShipEquipment(client.Commander.CommanderID, ownedShip.ID, 1)
	if err != nil {
		t.Fatalf("load ship equipment: %v", err)
	}
	if entry.EquipID != 2001 {
		t.Fatalf("expected equip unchanged")
	}
	if client.Commander.GetResourceCount(1) != 50 {
		t.Fatalf("expected gold unchanged")
	}
}

func TestTransformEquipmentOnShipMovesToBagWhenForbidden(t *testing.T) {
	client := setupTransformEquipmentTest(t)
	seedResource(t, 1)
	seedItem(t, 3001)
	seedCommanderGold(t, client.Commander.CommanderID, 200)
	seedCommanderItem(t, client.Commander.CommanderID, 3001, 2)

	ship := orm.Ship{TemplateID: 1001, Name: "Ship", EnglishName: "Ship", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	execAnswerExternalTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(ship.TemplateID), ship.Name, ship.EnglishName, int64(ship.RarityID), int64(ship.Star), int64(ship.Type), int64(ship.Nationality), int64(ship.BuildTime))
	seedShipEquipConfig(t, 1001, `{"id":1001,"equip_1":[1],"equip_2":[],"equip_3":[],"equip_4":[],"equip_5":[],"equip_id_1":2001,"equip_id_2":0,"equip_id_3":0}`)
	seedEquipment(t, 2001, 1, 0)
	seedEquipment(t, 2002, 2, 0)
	seedEquipUpgradeData(t, 9001, 2001, 2002, 100, [][]uint32{{3001, 2}})

	ownedShip, err := client.Commander.AddShip(1001)
	if err != nil {
		t.Fatalf("add ship: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	resp := sendCS14013(t, client, ownedShip.ID, 1, 9001)
	if resp.GetResult() != 0 {
		t.Fatalf("expected success")
	}
	entry, err := orm.GetOwnedShipEquipment(client.Commander.CommanderID, ownedShip.ID, 1)
	if err != nil {
		t.Fatalf("load ship equipment: %v", err)
	}
	if entry.EquipID != 0 {
		t.Fatalf("expected slot to be cleared")
	}
	owned := client.Commander.GetOwnedEquipment(2002)
	if owned == nil || owned.Count != 1 {
		t.Fatalf("expected upgraded equipment to be added to bag")
	}
}

func TestTransformEquipmentOnShipFailsInsufficientMaterial(t *testing.T) {
	client := setupTransformEquipmentTest(t)
	seedResource(t, 1)
	seedItem(t, 3001)
	seedCommanderGold(t, client.Commander.CommanderID, 200)
	seedCommanderItem(t, client.Commander.CommanderID, 3001, 1)

	ship := orm.Ship{TemplateID: 1001, Name: "Ship", EnglishName: "Ship", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	execAnswerExternalTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(ship.TemplateID), ship.Name, ship.EnglishName, int64(ship.RarityID), int64(ship.Star), int64(ship.Type), int64(ship.Nationality), int64(ship.BuildTime))
	seedShipEquipConfig(t, 1001, `{"id":1001,"equip_1":[1],"equip_2":[],"equip_3":[],"equip_4":[],"equip_5":[],"equip_id_1":2001,"equip_id_2":0,"equip_id_3":0}`)
	seedEquipment(t, 2001, 1, 0)
	seedEquipment(t, 2002, 1, 0)
	seedEquipUpgradeData(t, 9001, 2001, 2002, 100, [][]uint32{{3001, 2}})

	ownedShip, err := client.Commander.AddShip(1001)
	if err != nil {
		t.Fatalf("add ship: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}
	goldBefore := client.Commander.GetResourceCount(1)
	itemBefore := client.Commander.GetItemCount(3001)

	resp := sendCS14013(t, client, ownedShip.ID, 1, 9001)
	if resp.GetResult() == 0 {
		t.Fatalf("expected failure")
	}
	entry, err := orm.GetOwnedShipEquipment(client.Commander.CommanderID, ownedShip.ID, 1)
	if err != nil {
		t.Fatalf("load ship equipment: %v", err)
	}
	if entry.EquipID != 2001 {
		t.Fatalf("expected equip unchanged")
	}
	if client.Commander.GetResourceCount(1) != goldBefore {
		t.Fatalf("expected gold unchanged")
	}
	if client.Commander.GetItemCount(3001) != itemBefore {
		t.Fatalf("expected items unchanged")
	}
}
