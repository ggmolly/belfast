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
	commander := orm.Commander{CommanderID: 701, AccountID: 701, Name: "Upgrade Star Tester"}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
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
	if err := orm.GormDB.Create(&mainShip).Error; err != nil {
		t.Fatalf("create main ship: %v", err)
	}
	materialShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 2001, Level: 1}
	if err := orm.GormDB.Create(&materialShip).Error; err != nil {
		t.Fatalf("create material ship: %v", err)
	}
	equipEntry := orm.OwnedShipEquipment{OwnerID: client.Commander.CommanderID, ShipID: materialShip.ID, Pos: 1, EquipID: 3001, SkinID: 0}
	if err := orm.GormDB.Create(&equipEntry).Error; err != nil {
		t.Fatalf("seed ship equipment: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 1, Amount: 1000}).Error; err != nil {
		t.Fatalf("seed gold: %v", err)
	}
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: 18001, Count: 2}).Error; err != nil {
		t.Fatalf("seed item: %v", err)
	}
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

	var updated orm.OwnedShip
	if err := orm.GormDB.Where("owner_id = ? AND id = ?", client.Commander.CommanderID, mainShip.ID).First(&updated).Error; err != nil {
		t.Fatalf("load updated ship: %v", err)
	}
	if updated.ShipID != 1002 {
		t.Fatalf("expected ship_id 1002, got %d", updated.ShipID)
	}
	if updated.MaxLevel != 80 {
		t.Fatalf("expected max_level 80, got %d", updated.MaxLevel)
	}
	var gold orm.OwnedResource
	if err := orm.GormDB.Where("commander_id = ? AND resource_id = ?", client.Commander.CommanderID, 1).First(&gold).Error; err != nil {
		t.Fatalf("load gold: %v", err)
	}
	if gold.Amount != 700 {
		t.Fatalf("expected gold 700, got %d", gold.Amount)
	}
	var item orm.CommanderItem
	if err := orm.GormDB.Where("commander_id = ? AND item_id = ?", client.Commander.CommanderID, 18001).First(&item).Error; err != nil {
		t.Fatalf("load item: %v", err)
	}
	if item.Count != 0 {
		t.Fatalf("expected item count 0, got %d", item.Count)
	}
	var materialCheck orm.OwnedShip
	if err := orm.GormDB.Where("owner_id = ? AND id = ?", client.Commander.CommanderID, materialShip.ID).First(&materialCheck).Error; err == nil {
		t.Fatalf("expected material ship to be deleted")
	}
	var bag orm.OwnedEquipment
	if err := orm.GormDB.Where("commander_id = ? AND equipment_id = ?", client.Commander.CommanderID, 3001).First(&bag).Error; err != nil {
		t.Fatalf("load owned equipment: %v", err)
	}
	if bag.Count != 1 {
		t.Fatalf("expected equipment count 1, got %d", bag.Count)
	}
}

func TestUpgradeStarMaterialMismatch(t *testing.T) {
	client := setupUpgradeStarTest(t)
	seedBreakoutTemplate(t, 1001, 10, 70)
	seedBreakoutTemplate(t, 1002, 10, 80)
	seedBreakoutTemplate(t, 2001, 99, 70)
	seedBreakoutConfig(t, 1001, 1002, 10, 300, 10, 1, "[]")
	mainShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 15, MaxLevel: 70}
	if err := orm.GormDB.Create(&mainShip).Error; err != nil {
		t.Fatalf("create main ship: %v", err)
	}
	materialShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 2001, Level: 1}
	if err := orm.GormDB.Create(&materialShip).Error; err != nil {
		t.Fatalf("create material ship: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 1, Amount: 1000}).Error; err != nil {
		t.Fatalf("seed gold: %v", err)
	}
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
	var updated orm.OwnedShip
	if err := orm.GormDB.Where("owner_id = ? AND id = ?", client.Commander.CommanderID, mainShip.ID).First(&updated).Error; err != nil {
		t.Fatalf("load updated ship: %v", err)
	}
	if updated.ShipID != 1001 {
		t.Fatalf("expected ship_id 1001, got %d", updated.ShipID)
	}
	var materialCheck orm.OwnedShip
	if err := orm.GormDB.Where("owner_id = ? AND id = ?", client.Commander.CommanderID, materialShip.ID).First(&materialCheck).Error; err != nil {
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
	if err := orm.GormDB.Create(&mainShip).Error; err != nil {
		t.Fatalf("create main ship: %v", err)
	}
	materialShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 2001, Level: 1}
	if err := orm.GormDB.Create(&materialShip).Error; err != nil {
		t.Fatalf("create material ship: %v", err)
	}
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
	var materialCheck orm.OwnedShip
	if err := orm.GormDB.Where("owner_id = ? AND id = ?", client.Commander.CommanderID, materialShip.ID).First(&materialCheck).Error; err != nil {
		t.Fatalf("expected material ship to remain: %v", err)
	}
}
