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
	clearEquipTable(t, &orm.OwnedShipEquipment{})
	clearEquipTable(t, &orm.OwnedShip{})
	clearEquipTable(t, &orm.OwnedEquipment{})
	clearEquipTable(t, &orm.CommanderItem{})
	clearEquipTable(t, &orm.CommanderMiscItem{})
	clearEquipTable(t, &orm.OwnedResource{})
	clearEquipTable(t, &orm.Equipment{})
	clearEquipTable(t, &orm.Ship{})
	clearEquipTable(t, &orm.Resource{})
	clearEquipTable(t, &orm.Item{})
	clearEquipTable(t, &orm.ConfigEntry{})
	clearEquipTable(t, &orm.Commander{})

	commander := orm.Commander{CommanderID: 701, AccountID: 701, Name: "Transform Tester"}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func seedResource(t *testing.T, id uint32) {
	t.Helper()
	resource := orm.Resource{ID: id, Name: fmt.Sprintf("res-%d", id)}
	if err := orm.GormDB.Create(&resource).Error; err != nil {
		t.Fatalf("seed resource: %v", err)
	}
}

func seedItem(t *testing.T, id uint32) {
	t.Helper()
	item := orm.Item{ID: id, Name: fmt.Sprintf("item-%d", id), Rarity: 1, ShopID: -2, Type: 1, VirtualType: 0}
	if err := orm.GormDB.Create(&item).Error; err != nil {
		t.Fatalf("seed item: %v", err)
	}
}

func seedCommanderGold(t *testing.T, commanderID uint32, amount uint32) {
	t.Helper()
	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: commanderID, ResourceID: 1, Amount: amount}).Error; err != nil {
		t.Fatalf("seed gold: %v", err)
	}
}

func seedCommanderItem(t *testing.T, commanderID uint32, itemID uint32, count uint32) {
	t.Helper()
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: commanderID, ItemID: itemID, Count: count}).Error; err != nil {
		t.Fatalf("seed commander item: %v", err)
	}
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
	if err := orm.GormDB.Create(&entry).Error; err != nil {
		t.Fatalf("seed equip upgrade data: %v", err)
	}
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
	if err := orm.GormDB.Create(&equip).Error; err != nil {
		t.Fatalf("seed equipment: %v", err)
	}
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
	if err := orm.GormDB.Create(&ship).Error; err != nil {
		t.Fatalf("create ship: %v", err)
	}
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
	var entry orm.OwnedShipEquipment
	if err := orm.GormDB.Where("owner_id = ? AND ship_id = ? AND pos = ?", client.Commander.CommanderID, ownedShip.ID, 1).First(&entry).Error; err != nil {
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
	if err := orm.GormDB.Create(&ship).Error; err != nil {
		t.Fatalf("create ship: %v", err)
	}
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
	var entry orm.OwnedShipEquipment
	if err := orm.GormDB.Where("owner_id = ? AND ship_id = ? AND pos = ?", client.Commander.CommanderID, ownedShip.ID, 1).First(&entry).Error; err != nil {
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
	if err := orm.GormDB.Create(&ship).Error; err != nil {
		t.Fatalf("create ship: %v", err)
	}
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
	var entry orm.OwnedShipEquipment
	if err := orm.GormDB.Where("owner_id = ? AND ship_id = ? AND pos = ?", client.Commander.CommanderID, ownedShip.ID, 1).First(&entry).Error; err != nil {
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
	if err := orm.GormDB.Create(&ship).Error; err != nil {
		t.Fatalf("create ship: %v", err)
	}
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
	var entry orm.OwnedShipEquipment
	if err := orm.GormDB.Where("owner_id = ? AND ship_id = ? AND pos = ?", client.Commander.CommanderID, ownedShip.ID, 1).First(&entry).Error; err != nil {
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
	if err := orm.GormDB.Create(&ship).Error; err != nil {
		t.Fatalf("create ship: %v", err)
	}
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
	var entry orm.OwnedShipEquipment
	if err := orm.GormDB.Where("owner_id = ? AND ship_id = ? AND pos = ?", client.Commander.CommanderID, ownedShip.ID, 1).First(&entry).Error; err != nil {
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
