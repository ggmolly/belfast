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

func setupUpgradeEquipmentOnShip14002Test(t *testing.T, commanderID uint32) *connection.Client {
	t.Helper()

	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearEquipTable(t, &orm.OwnedShipEquipment{})
	clearEquipTable(t, &orm.OwnedShip{})
	clearEquipTable(t, &orm.OwnedResource{})
	clearEquipTable(t, &orm.CommanderItem{})
	clearEquipTable(t, &orm.CommanderMiscItem{})
	clearEquipTable(t, &orm.Equipment{})
	clearEquipTable(t, &orm.Ship{})
	clearEquipTable(t, &orm.Commander{})

	commander := orm.Commander{CommanderID: commanderID, AccountID: commanderID, Name: "Upgrade Equip Ship"}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func TestUpgradeEquipmentOnShip14002SuccessUpdatesSlotAndChargesCosts(t *testing.T) {
	client := setupUpgradeEquipmentOnShip14002Test(t, 9014002)

	shipTemplate := orm.Ship{TemplateID: 1001, Name: "Ship", EnglishName: "Ship", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	if err := orm.GormDB.Create(&shipTemplate).Error; err != nil {
		t.Fatalf("create ship: %v", err)
	}
	ownedShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1001, ID: 50001, Level: 1, MaxLevel: 50}
	if err := orm.GormDB.Create(&ownedShip).Error; err != nil {
		t.Fatalf("create owned ship: %v", err)
	}

	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 1, Amount: 100}).Error; err != nil {
		t.Fatalf("create gold resource: %v", err)
	}
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: 200, Count: 3}).Error; err != nil {
		t.Fatalf("create item: %v", err)
	}

	baseID := uint32(995000)
	upgradedID := uint32(995001)
	createTestEquipment(t, baseID, upgradedID, 10, json.RawMessage(`[[200,1]]`))
	createTestEquipment(t, upgradedID, 0, 0, nil)

	if err := orm.GormDB.Create(&orm.OwnedShipEquipment{OwnerID: client.Commander.CommanderID, ShipID: ownedShip.ID, Pos: 1, EquipID: baseID, SkinID: 123}).Error; err != nil {
		t.Fatalf("create ship equipment: %v", err)
	}

	payload := &protobuf.CS_14002{ShipId: proto.Uint32(ownedShip.ID), Pos: proto.Uint32(1), Lv: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.UpgradeEquipmentOnShip14002(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	resp := &protobuf.SC_14003{}
	decodePacket(t, client, 14003, resp)
	if resp.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", resp.GetResult())
	}

	var updated orm.OwnedShipEquipment
	if err := orm.GormDB.Where("owner_id = ? AND ship_id = ? AND pos = ?", client.Commander.CommanderID, ownedShip.ID, 1).First(&updated).Error; err != nil {
		t.Fatalf("load ship equipment: %v", err)
	}
	if updated.EquipID != upgradedID {
		t.Fatalf("expected equip id %d, got %d", upgradedID, updated.EquipID)
	}
	if updated.SkinID != 123 {
		t.Fatalf("expected skin id 123 to remain, got %d", updated.SkinID)
	}

	var gold orm.OwnedResource
	if err := orm.GormDB.Where("commander_id = ? AND resource_id = ?", client.Commander.CommanderID, 1).First(&gold).Error; err != nil {
		t.Fatalf("load gold: %v", err)
	}
	if gold.Amount != 90 {
		t.Fatalf("expected gold 90, got %d", gold.Amount)
	}

	var item orm.CommanderItem
	if err := orm.GormDB.Where("commander_id = ? AND item_id = ?", client.Commander.CommanderID, 200).First(&item).Error; err != nil {
		t.Fatalf("load item: %v", err)
	}
	if item.Count != 2 {
		t.Fatalf("expected item count 2, got %d", item.Count)
	}
}

func TestUpgradeEquipmentOnShip14002ShipIDZero(t *testing.T) {
	client := &connection.Client{Commander: &orm.Commander{CommanderID: 9014010, AccountID: 9014010, Name: "shipid-zero"}}

	payload := &protobuf.CS_14002{ShipId: proto.Uint32(0), Pos: proto.Uint32(1), Lv: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := answer.UpgradeEquipmentOnShip14002(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	resp := &protobuf.SC_14003{}
	decodePacket(t, client, 14003, resp)
	if resp.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}

func TestUpgradeEquipmentOnShip14002PosZero(t *testing.T) {
	client := &connection.Client{Commander: &orm.Commander{CommanderID: 9014011, AccountID: 9014011, Name: "pos-zero"}}

	payload := &protobuf.CS_14002{ShipId: proto.Uint32(1), Pos: proto.Uint32(0), Lv: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := answer.UpgradeEquipmentOnShip14002(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	resp := &protobuf.SC_14003{}
	decodePacket(t, client, 14003, resp)
	if resp.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}

func TestUpgradeEquipmentOnShip14002LvZero(t *testing.T) {
	client := &connection.Client{Commander: &orm.Commander{CommanderID: 9014012, AccountID: 9014012, Name: "lv-zero"}}

	payload := &protobuf.CS_14002{ShipId: proto.Uint32(1), Pos: proto.Uint32(1), Lv: proto.Uint32(0)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := answer.UpgradeEquipmentOnShip14002(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	resp := &protobuf.SC_14003{}
	decodePacket(t, client, 14003, resp)
	if resp.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}

func TestUpgradeEquipmentOnShip14002NonOwnedShipFails(t *testing.T) {
	client := setupUpgradeEquipmentOnShip14002Test(t, 9014003)

	payload := &protobuf.CS_14002{ShipId: proto.Uint32(99999), Pos: proto.Uint32(1), Lv: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.UpgradeEquipmentOnShip14002(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	resp := &protobuf.SC_14003{}
	decodePacket(t, client, 14003, resp)
	if resp.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}

func TestUpgradeEquipmentOnShip14002EmptySlotFails(t *testing.T) {
	client := setupUpgradeEquipmentOnShip14002Test(t, 9014004)

	shipTemplate := orm.Ship{TemplateID: 1001, Name: "Ship", EnglishName: "Ship", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	if err := orm.GormDB.Create(&shipTemplate).Error; err != nil {
		t.Fatalf("create ship: %v", err)
	}
	ownedShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1001, ID: 50002, Level: 1, MaxLevel: 50}
	if err := orm.GormDB.Create(&ownedShip).Error; err != nil {
		t.Fatalf("create owned ship: %v", err)
	}

	payload := &protobuf.CS_14002{ShipId: proto.Uint32(ownedShip.ID), Pos: proto.Uint32(1), Lv: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.UpgradeEquipmentOnShip14002(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	resp := &protobuf.SC_14003{}
	decodePacket(t, client, 14003, resp)
	if resp.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}

func TestUpgradeEquipmentOnShip14002NotEnoughGoldDoesNotMutate(t *testing.T) {
	client := setupUpgradeEquipmentOnShip14002Test(t, 9014005)

	shipTemplate := orm.Ship{TemplateID: 1001, Name: "Ship", EnglishName: "Ship", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	if err := orm.GormDB.Create(&shipTemplate).Error; err != nil {
		t.Fatalf("create ship: %v", err)
	}
	ownedShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1001, ID: 50003, Level: 1, MaxLevel: 50}
	if err := orm.GormDB.Create(&ownedShip).Error; err != nil {
		t.Fatalf("create owned ship: %v", err)
	}

	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 1, Amount: 5}).Error; err != nil {
		t.Fatalf("create gold: %v", err)
	}
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: 200, Count: 3}).Error; err != nil {
		t.Fatalf("create item: %v", err)
	}

	baseID := uint32(996000)
	upgradedID := uint32(996001)
	createTestEquipment(t, baseID, upgradedID, 10, json.RawMessage(`[[200,1]]`))
	createTestEquipment(t, upgradedID, 0, 0, nil)
	if err := orm.GormDB.Create(&orm.OwnedShipEquipment{OwnerID: client.Commander.CommanderID, ShipID: ownedShip.ID, Pos: 1, EquipID: baseID, SkinID: 0}).Error; err != nil {
		t.Fatalf("create ship equipment: %v", err)
	}

	payload := &protobuf.CS_14002{ShipId: proto.Uint32(ownedShip.ID), Pos: proto.Uint32(1), Lv: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.UpgradeEquipmentOnShip14002(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	resp := &protobuf.SC_14003{}
	decodePacket(t, client, 14003, resp)
	if resp.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}

	var entry orm.OwnedShipEquipment
	if err := orm.GormDB.Where("owner_id = ? AND ship_id = ? AND pos = ?", client.Commander.CommanderID, ownedShip.ID, 1).First(&entry).Error; err != nil {
		t.Fatalf("load ship equipment: %v", err)
	}
	if entry.EquipID != baseID {
		t.Fatalf("expected equip id %d, got %d", baseID, entry.EquipID)
	}
}

func TestUpgradeEquipmentOnShip14002NotEnoughItemsDoesNotMutate(t *testing.T) {
	client := setupUpgradeEquipmentOnShip14002Test(t, 9014006)

	shipTemplate := orm.Ship{TemplateID: 1001, Name: "Ship", EnglishName: "Ship", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	if err := orm.GormDB.Create(&shipTemplate).Error; err != nil {
		t.Fatalf("create ship: %v", err)
	}
	ownedShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1001, ID: 50004, Level: 1, MaxLevel: 50}
	if err := orm.GormDB.Create(&ownedShip).Error; err != nil {
		t.Fatalf("create owned ship: %v", err)
	}

	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 1, Amount: 100}).Error; err != nil {
		t.Fatalf("create gold: %v", err)
	}
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: client.Commander.CommanderID, ItemID: 200, Count: 0}).Error; err != nil {
		t.Fatalf("create item: %v", err)
	}

	baseID := uint32(997000)
	upgradedID := uint32(997001)
	createTestEquipment(t, baseID, upgradedID, 10, json.RawMessage(`[[200,1]]`))
	createTestEquipment(t, upgradedID, 0, 0, nil)
	if err := orm.GormDB.Create(&orm.OwnedShipEquipment{OwnerID: client.Commander.CommanderID, ShipID: ownedShip.ID, Pos: 1, EquipID: baseID, SkinID: 0}).Error; err != nil {
		t.Fatalf("create ship equipment: %v", err)
	}

	payload := &protobuf.CS_14002{ShipId: proto.Uint32(ownedShip.ID), Pos: proto.Uint32(1), Lv: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.UpgradeEquipmentOnShip14002(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	resp := &protobuf.SC_14003{}
	decodePacket(t, client, 14003, resp)
	if resp.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}

	var entry orm.OwnedShipEquipment
	if err := orm.GormDB.Where("owner_id = ? AND ship_id = ? AND pos = ?", client.Commander.CommanderID, ownedShip.ID, 1).First(&entry).Error; err != nil {
		t.Fatalf("load ship equipment: %v", err)
	}
	if entry.EquipID != baseID {
		t.Fatalf("expected equip id %d, got %d", baseID, entry.EquipID)
	}
}
