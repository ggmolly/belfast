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
)

func setupDestroyEquipmentsTest(t *testing.T) *connection.Client {
	t.Helper()
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearEquipTable(t, &orm.OwnedEquipment{})
	clearEquipTable(t, &orm.Equipment{})
	clearEquipTable(t, &orm.CommanderItem{})
	clearEquipTable(t, &orm.OwnedResource{})
	clearEquipTable(t, &orm.Commander{})
	commander := orm.Commander{CommanderID: 1001, AccountID: 1001, Name: "Destroy Equipments Tester"}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func TestDestroyEquipmentsSuccessAwardsRewards(t *testing.T) {
	client := setupDestroyEquipmentsTest(t)
	if err := orm.GormDB.Create(&orm.Equipment{ID: 2001, DestroyGold: 10, DestroyItem: json.RawMessage(`[[300,1],[301,2]]`), ShipTypeForbidden: json.RawMessage(`[]`)}).Error; err != nil {
		t.Fatalf("seed equipment: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedEquipment{CommanderID: client.Commander.CommanderID, EquipmentID: 2001, Count: 5}).Error; err != nil {
		t.Fatalf("seed owned equipment: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 1, Amount: 50}).Error; err != nil {
		t.Fatalf("seed gold: %v", err)
	}
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: 300, Count: 7}).Error; err != nil {
		t.Fatalf("seed item 300: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	payload := protobuf.CS_14008{EquipList: []*protobuf.EQUIPINFO{{Id: proto.Uint32(2001), Count: proto.Uint32(3)}}}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.DestroyEquipments(&buf, client); err != nil {
		t.Fatalf("DestroyEquipments failed: %v", err)
	}
	response := &protobuf.SC_14009{}
	decodePacket(t, client, 14009, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}

	if count := loadOwnedEquipmentCount(t, client.Commander.CommanderID, 2001); count != 2 {
		t.Fatalf("expected owned equipment count 2, got %d", count)
	}
	if amount := loadResourceCount(t, client.Commander.CommanderID, 1); amount != 80 {
		t.Fatalf("expected gold 80, got %d", amount)
	}
	if count := loadItemCount(t, client.Commander.CommanderID, 300); count != 10 {
		t.Fatalf("expected item 300 count 10, got %d", count)
	}
	if count := loadItemCount(t, client.Commander.CommanderID, 301); count != 6 {
		t.Fatalf("expected item 301 count 6, got %d", count)
	}
}

func TestDestroyEquipmentsInsufficientEquipmentDoesNotMutate(t *testing.T) {
	client := setupDestroyEquipmentsTest(t)
	if err := orm.GormDB.Create(&orm.Equipment{ID: 2002, DestroyGold: 9, DestroyItem: json.RawMessage(`[[302,1]]`), ShipTypeForbidden: json.RawMessage(`[]`)}).Error; err != nil {
		t.Fatalf("seed equipment: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedEquipment{CommanderID: client.Commander.CommanderID, EquipmentID: 2002, Count: 1}).Error; err != nil {
		t.Fatalf("seed owned equipment: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 1, Amount: 5}).Error; err != nil {
		t.Fatalf("seed gold: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	beforeEquip := loadOwnedEquipmentCount(t, client.Commander.CommanderID, 2002)
	beforeGold := loadResourceCount(t, client.Commander.CommanderID, 1)

	payload := protobuf.CS_14008{EquipList: []*protobuf.EQUIPINFO{{Id: proto.Uint32(2002), Count: proto.Uint32(2)}}}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.DestroyEquipments(&buf, client); err != nil {
		t.Fatalf("DestroyEquipments failed: %v", err)
	}
	response := &protobuf.SC_14009{}
	decodePacket(t, client, 14009, response)
	if response.GetResult() != 2 {
		t.Fatalf("expected result 2, got %d", response.GetResult())
	}

	if after := loadOwnedEquipmentCount(t, client.Commander.CommanderID, 2002); after != beforeEquip {
		t.Fatalf("expected equipment unchanged")
	}
	if after := loadResourceCount(t, client.Commander.CommanderID, 1); after != beforeGold {
		t.Fatalf("expected gold unchanged")
	}
	if count := loadItemCount(t, client.Commander.CommanderID, 302); count != 0 {
		t.Fatalf("expected no reward items")
	}
}

func TestDestroyEquipmentsUnknownTemplateDoesNotMutate(t *testing.T) {
	client := setupDestroyEquipmentsTest(t)
	if err := orm.GormDB.Create(&orm.OwnedEquipment{CommanderID: client.Commander.CommanderID, EquipmentID: 9001, Count: 2}).Error; err != nil {
		t.Fatalf("seed owned equipment: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	beforeEquip := loadOwnedEquipmentCount(t, client.Commander.CommanderID, 9001)

	payload := protobuf.CS_14008{EquipList: []*protobuf.EQUIPINFO{{Id: proto.Uint32(9001), Count: proto.Uint32(1)}}}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.DestroyEquipments(&buf, client); err != nil {
		t.Fatalf("DestroyEquipments failed: %v", err)
	}
	response := &protobuf.SC_14009{}
	decodePacket(t, client, 14009, response)
	if response.GetResult() != 3 {
		t.Fatalf("expected result 3, got %d", response.GetResult())
	}
	if after := loadOwnedEquipmentCount(t, client.Commander.CommanderID, 9001); after != beforeEquip {
		t.Fatalf("expected equipment unchanged")
	}
}

func TestDestroyEquipmentsMultiEntryAppliesAll(t *testing.T) {
	client := setupDestroyEquipmentsTest(t)
	if err := orm.GormDB.Create(&orm.Equipment{ID: 2100, DestroyGold: 2, DestroyItem: json.RawMessage(`[[400,1]]`), ShipTypeForbidden: json.RawMessage(`[]`)}).Error; err != nil {
		t.Fatalf("seed equipment 2100: %v", err)
	}
	if err := orm.GormDB.Create(&orm.Equipment{ID: 2101, DestroyGold: 3, DestroyItem: json.RawMessage(`[[401,2]]`), ShipTypeForbidden: json.RawMessage(`[]`)}).Error; err != nil {
		t.Fatalf("seed equipment 2101: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedEquipment{CommanderID: client.Commander.CommanderID, EquipmentID: 2100, Count: 4}).Error; err != nil {
		t.Fatalf("seed owned equipment 2100: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedEquipment{CommanderID: client.Commander.CommanderID, EquipmentID: 2101, Count: 1}).Error; err != nil {
		t.Fatalf("seed owned equipment 2101: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 1, Amount: 0}).Error; err != nil {
		t.Fatalf("seed gold: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	payload := protobuf.CS_14008{EquipList: []*protobuf.EQUIPINFO{
		{Id: proto.Uint32(2100), Count: proto.Uint32(2)},
		{Id: proto.Uint32(2101), Count: proto.Uint32(1)},
	}}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.DestroyEquipments(&buf, client); err != nil {
		t.Fatalf("DestroyEquipments failed: %v", err)
	}
	response := &protobuf.SC_14009{}
	decodePacket(t, client, 14009, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}

	if count := loadOwnedEquipmentCount(t, client.Commander.CommanderID, 2100); count != 2 {
		t.Fatalf("expected equip 2100 count 2, got %d", count)
	}
	if count := loadOwnedEquipmentCount(t, client.Commander.CommanderID, 2101); count != 0 {
		t.Fatalf("expected equip 2101 removed")
	}
	if amount := loadResourceCount(t, client.Commander.CommanderID, 1); amount != 7 {
		t.Fatalf("expected gold 7, got %d", amount)
	}
	if count := loadItemCount(t, client.Commander.CommanderID, 400); count != 2 {
		t.Fatalf("expected item 400 count 2, got %d", count)
	}
	if count := loadItemCount(t, client.Commander.CommanderID, 401); count != 2 {
		t.Fatalf("expected item 401 count 2, got %d", count)
	}
}

func TestDestroyEquipmentsZeroCountDoesNotMutate(t *testing.T) {
	client := setupDestroyEquipmentsTest(t)
	if err := orm.GormDB.Create(&orm.Equipment{ID: 2200, DestroyGold: 1, DestroyItem: json.RawMessage(`[[500,1]]`), ShipTypeForbidden: json.RawMessage(`[]`)}).Error; err != nil {
		t.Fatalf("seed equipment: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedEquipment{CommanderID: client.Commander.CommanderID, EquipmentID: 2200, Count: 1}).Error; err != nil {
		t.Fatalf("seed owned equipment: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 1, Amount: 0}).Error; err != nil {
		t.Fatalf("seed gold: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}

	beforeEquip := loadOwnedEquipmentCount(t, client.Commander.CommanderID, 2200)

	payload := protobuf.CS_14008{EquipList: []*protobuf.EQUIPINFO{{Id: proto.Uint32(2200), Count: proto.Uint32(0)}}}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.DestroyEquipments(&buf, client); err != nil {
		t.Fatalf("DestroyEquipments failed: %v", err)
	}
	response := &protobuf.SC_14009{}
	decodePacket(t, client, 14009, response)
	if response.GetResult() != 1 {
		t.Fatalf("expected result 1, got %d", response.GetResult())
	}
	if after := loadOwnedEquipmentCount(t, client.Commander.CommanderID, 2200); after != beforeEquip {
		t.Fatalf("expected equipment unchanged")
	}
	if amount := loadResourceCount(t, client.Commander.CommanderID, 1); amount != 0 {
		t.Fatalf("expected gold unchanged")
	}
	if count := loadItemCount(t, client.Commander.CommanderID, 500); count != 0 {
		t.Fatalf("expected no item rewards")
	}
}
