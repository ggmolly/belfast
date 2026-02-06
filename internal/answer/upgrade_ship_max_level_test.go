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

	if err := orm.GormDB.Create(&orm.Ship{TemplateID: 1001, Name: "Test Ship", EnglishName: "Test Ship", RarityID: 3, Star: 1, Type: 1, Nationality: 1}).Error; err != nil {
		t.Fatalf("seed ship template: %v", err)
	}
	owned := orm.OwnedShip{ID: 101, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 100, MaxLevel: 100, Exp: 0, SurplusExp: 0, Energy: 150}
	if err := orm.GormDB.Create(&owned).Error; err != nil {
		t.Fatalf("seed owned ship: %v", err)
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

	var updated orm.OwnedShip
	if err := orm.GormDB.Where("owner_id = ? AND id = ?", client.Commander.CommanderID, owned.ID).First(&updated).Error; err != nil {
		t.Fatalf("load updated ship: %v", err)
	}
	if updated.MaxLevel != 105 {
		t.Fatalf("expected max_level 105, got %d", updated.MaxLevel)
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

	if err := orm.GormDB.Create(&orm.Ship{TemplateID: 1001, Name: "Test Ship", EnglishName: "Test Ship", RarityID: 3, Star: 1, Type: 1, Nationality: 1}).Error; err != nil {
		t.Fatalf("seed ship template: %v", err)
	}
	owned := orm.OwnedShip{ID: 101, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 100, MaxLevel: 100, Exp: 0, SurplusExp: 0, Energy: 150}
	if err := orm.GormDB.Create(&owned).Error; err != nil {
		t.Fatalf("seed owned ship: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 1, Amount: 100}).Error; err != nil {
		t.Fatalf("seed gold: %v", err)
	}
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: 18001, Count: 2}).Error; err != nil {
		t.Fatalf("seed item: %v", err)
	}
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

	var updated orm.OwnedShip
	if err := orm.GormDB.Where("owner_id = ? AND id = ?", client.Commander.CommanderID, owned.ID).First(&updated).Error; err != nil {
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

	if err := orm.GormDB.Create(&orm.Ship{TemplateID: 1001, Name: "Test Ship", EnglishName: "Test Ship", RarityID: 3, Star: 1, Type: 1, Nationality: 1}).Error; err != nil {
		t.Fatalf("seed ship template: %v", err)
	}
	owned := orm.OwnedShip{ID: 101, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 99, MaxLevel: 100, Exp: 0, SurplusExp: 0, Energy: 150}
	if err := orm.GormDB.Create(&owned).Error; err != nil {
		t.Fatalf("seed owned ship: %v", err)
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

	if err := orm.GormDB.Create(&orm.Ship{TemplateID: 1001, Name: "Test Ship", EnglishName: "Test Ship", RarityID: 3, Star: 1, Type: 1, Nationality: 1}).Error; err != nil {
		t.Fatalf("seed ship template: %v", err)
	}
	owned := orm.OwnedShip{ID: 101, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 100, MaxLevel: 100, Exp: 0, SurplusExp: 35, Energy: 150}
	if err := orm.GormDB.Create(&owned).Error; err != nil {
		t.Fatalf("seed owned ship: %v", err)
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

	var updated orm.OwnedShip
	if err := orm.GormDB.Where("owner_id = ? AND id = ?", client.Commander.CommanderID, owned.ID).First(&updated).Error; err != nil {
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
