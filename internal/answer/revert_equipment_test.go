package answer_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func setupRevertEquipmentTest(t *testing.T) *connection.Client {
	t.Helper()
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearEquipTable(t, &orm.OwnedEquipment{})
	clearEquipTable(t, &orm.Equipment{})
	clearEquipTable(t, &orm.CommanderItem{})
	clearEquipTable(t, &orm.CommanderMiscItem{})
	clearEquipTable(t, &orm.OwnedResource{})
	clearEquipTable(t, &orm.Commander{})
	commander := orm.Commander{CommanderID: 901, AccountID: 901, Name: "Revert Equipment Tester"}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func seedRevertEquipmentChain(t *testing.T) {
	t.Helper()
	entries := []orm.Equipment{
		{ID: 500, Prev: 0, Level: 1, TransUseGold: 10, TransUseItem: json.RawMessage(`[[200,1]]`), ShipTypeForbidden: json.RawMessage(`[]`)},
		{ID: 501, Prev: 500, Level: 2, TransUseGold: 20, TransUseItem: json.RawMessage(`[[200,2],[201,1]]`), ShipTypeForbidden: json.RawMessage(`[]`)},
		{ID: 502, Prev: 501, Level: 3, TransUseGold: 30, TransUseItem: json.RawMessage(`[[200,3]]`), ShipTypeForbidden: json.RawMessage(`[]`)},
	}
	for _, entry := range entries {
		if err := orm.GormDB.Create(&entry).Error; err != nil {
			t.Fatalf("seed equipment %d: %v", entry.ID, err)
		}
	}
}

func loadOwnedEquipmentCount(t *testing.T, commanderID uint32, equipmentID uint32) uint32 {
	t.Helper()
	var entry orm.OwnedEquipment
	err := orm.GormDB.Where("commander_id = ? AND equipment_id = ?", commanderID, equipmentID).First(&entry).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0
		}
		t.Fatalf("load owned equipment: %v", err)
	}
	return entry.Count
}

func loadItemCount(t *testing.T, commanderID uint32, itemID uint32) uint32 {
	t.Helper()
	var entry orm.CommanderItem
	err := orm.GormDB.Where("commander_id = ? AND item_id = ?", commanderID, itemID).First(&entry).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0
		}
		t.Fatalf("load item: %v", err)
	}
	return entry.Count
}

func loadResourceCount(t *testing.T, commanderID uint32, resourceID uint32) uint32 {
	t.Helper()
	var entry orm.OwnedResource
	err := orm.GormDB.Where("commander_id = ? AND resource_id = ?", commanderID, resourceID).First(&entry).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0
		}
		t.Fatalf("load resource: %v", err)
	}
	return entry.Amount
}

func TestRevertEquipmentSuccess(t *testing.T) {
	client := setupRevertEquipmentTest(t)
	seedRevertEquipmentChain(t)
	if err := orm.GormDB.Create(&orm.OwnedEquipment{CommanderID: client.Commander.CommanderID, EquipmentID: 502, Count: 1}).Error; err != nil {
		t.Fatalf("seed owned equipment: %v", err)
	}
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: 15007, Count: 1}).Error; err != nil {
		t.Fatalf("seed revert item: %v", err)
	}
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: 200, Count: 10}).Error; err != nil {
		t.Fatalf("seed item 200: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 1, Amount: 100}).Error; err != nil {
		t.Fatalf("seed coins: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	payload := protobuf.CS_14010{EquipId: proto.Uint32(502)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.RevertEquipment(&buf, client); err != nil {
		t.Fatalf("RevertEquipment failed: %v", err)
	}
	response := &protobuf.SC_14011{}
	decodePacket(t, client, 14011, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected success result, got %d", response.GetResult())
	}

	if count := loadOwnedEquipmentCount(t, client.Commander.CommanderID, 502); count != 0 {
		t.Fatalf("expected equip 502 removed, got count %d", count)
	}
	if count := loadOwnedEquipmentCount(t, client.Commander.CommanderID, 500); count != 1 {
		t.Fatalf("expected root equip 500 count 1, got %d", count)
	}
	if count := loadItemCount(t, client.Commander.CommanderID, 15007); count != 0 {
		t.Fatalf("expected revert item consumed, got %d", count)
	}
	if count := loadItemCount(t, client.Commander.CommanderID, 200); count != 13 {
		t.Fatalf("expected item 200 refunded to 13, got %d", count)
	}
	if count := loadItemCount(t, client.Commander.CommanderID, 201); count != 1 {
		t.Fatalf("expected item 201 refunded to 1, got %d", count)
	}
	if amount := loadResourceCount(t, client.Commander.CommanderID, 1); amount != 130 {
		t.Fatalf("expected coins refunded to 130, got %d", amount)
	}
}

func TestRevertEquipmentMissingRevertItemDoesNotMutate(t *testing.T) {
	client := setupRevertEquipmentTest(t)
	seedRevertEquipmentChain(t)
	if err := orm.GormDB.Create(&orm.OwnedEquipment{CommanderID: client.Commander.CommanderID, EquipmentID: 502, Count: 1}).Error; err != nil {
		t.Fatalf("seed owned equipment: %v", err)
	}
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: 200, Count: 10}).Error; err != nil {
		t.Fatalf("seed item 200: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 1, Amount: 100}).Error; err != nil {
		t.Fatalf("seed coins: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	beforeEquip := loadOwnedEquipmentCount(t, client.Commander.CommanderID, 502)
	beforeCoins := loadResourceCount(t, client.Commander.CommanderID, 1)
	beforeItem200 := loadItemCount(t, client.Commander.CommanderID, 200)

	payload := protobuf.CS_14010{EquipId: proto.Uint32(502)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.RevertEquipment(&buf, client); err != nil {
		t.Fatalf("RevertEquipment failed: %v", err)
	}
	response := &protobuf.SC_14011{}
	decodePacket(t, client, 14011, response)
	if response.GetResult() == 0 {
		t.Fatalf("expected failure result")
	}

	if after := loadOwnedEquipmentCount(t, client.Commander.CommanderID, 502); after != beforeEquip {
		t.Fatalf("expected equip unchanged")
	}
	if after := loadOwnedEquipmentCount(t, client.Commander.CommanderID, 500); after != 0 {
		t.Fatalf("expected root equip not added")
	}
	if after := loadResourceCount(t, client.Commander.CommanderID, 1); after != beforeCoins {
		t.Fatalf("expected coins unchanged")
	}
	if after := loadItemCount(t, client.Commander.CommanderID, 200); after != beforeItem200 {
		t.Fatalf("expected refund items unchanged")
	}
}

func TestRevertEquipmentNotOwnedDoesNotMutate(t *testing.T) {
	client := setupRevertEquipmentTest(t)
	seedRevertEquipmentChain(t)
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: 15007, Count: 1}).Error; err != nil {
		t.Fatalf("seed revert item: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 1, Amount: 100}).Error; err != nil {
		t.Fatalf("seed coins: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	payload := protobuf.CS_14010{EquipId: proto.Uint32(502)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.RevertEquipment(&buf, client); err != nil {
		t.Fatalf("RevertEquipment failed: %v", err)
	}
	response := &protobuf.SC_14011{}
	decodePacket(t, client, 14011, response)
	if response.GetResult() == 0 {
		t.Fatalf("expected failure result")
	}
	if count := loadItemCount(t, client.Commander.CommanderID, 15007); count != 1 {
		t.Fatalf("expected revert item not consumed, got %d", count)
	}
}

func TestRevertEquipmentNotRevertableDoesNotMutate(t *testing.T) {
	client := setupRevertEquipmentTest(t)
	if err := orm.GormDB.Create(&orm.Equipment{ID: 600, Prev: 0, Level: 1, ShipTypeForbidden: json.RawMessage(`[]`)}).Error; err != nil {
		t.Fatalf("seed equipment: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedEquipment{CommanderID: client.Commander.CommanderID, EquipmentID: 600, Count: 1}).Error; err != nil {
		t.Fatalf("seed owned equipment: %v", err)
	}
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: 15007, Count: 1}).Error; err != nil {
		t.Fatalf("seed revert item: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	payload := protobuf.CS_14010{EquipId: proto.Uint32(600)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.RevertEquipment(&buf, client); err != nil {
		t.Fatalf("RevertEquipment failed: %v", err)
	}
	response := &protobuf.SC_14011{}
	decodePacket(t, client, 14011, response)
	if response.GetResult() == 0 {
		t.Fatalf("expected failure result")
	}
	if count := loadItemCount(t, client.Commander.CommanderID, 15007); count != 1 {
		t.Fatalf("expected revert item not consumed")
	}
	if count := loadOwnedEquipmentCount(t, client.Commander.CommanderID, 600); count != 1 {
		t.Fatalf("expected equipment not removed")
	}
}

func TestRevertEquipmentConfigMissingDoesNotMutate(t *testing.T) {
	client := setupRevertEquipmentTest(t)
	if err := orm.GormDB.Create(&orm.OwnedEquipment{CommanderID: client.Commander.CommanderID, EquipmentID: 700, Count: 1}).Error; err != nil {
		t.Fatalf("seed owned equipment: %v", err)
	}
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: 15007, Count: 1}).Error; err != nil {
		t.Fatalf("seed revert item: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	payload := protobuf.CS_14010{EquipId: proto.Uint32(700)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.RevertEquipment(&buf, client); err != nil {
		t.Fatalf("RevertEquipment failed: %v", err)
	}
	response := &protobuf.SC_14011{}
	decodePacket(t, client, 14011, response)
	if response.GetResult() == 0 {
		t.Fatalf("expected failure result")
	}
	if count := loadItemCount(t, client.Commander.CommanderID, 15007); count != 1 {
		t.Fatalf("expected revert item not consumed")
	}
	if count := loadOwnedEquipmentCount(t, client.Commander.CommanderID, 700); count != 1 {
		t.Fatalf("expected equipment not removed")
	}
}

func TestRevertEquipmentEquipIdZeroDoesNotMutate(t *testing.T) {
	client := setupRevertEquipmentTest(t)
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: 15007, Count: 1}).Error; err != nil {
		t.Fatalf("seed revert item: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	before := loadItemCount(t, client.Commander.CommanderID, 15007)
	payload := protobuf.CS_14010{EquipId: proto.Uint32(0)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.RevertEquipment(&buf, client); err != nil {
		t.Fatalf("RevertEquipment failed: %v", err)
	}
	response := &protobuf.SC_14011{}
	decodePacket(t, client, 14011, response)
	if response.GetResult() == 0 {
		t.Fatalf("expected failure result")
	}
	if after := loadItemCount(t, client.Commander.CommanderID, 15007); after != before {
		t.Fatalf("expected revert item unchanged")
	}
}
