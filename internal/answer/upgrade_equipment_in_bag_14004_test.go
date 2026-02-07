package answer_test

import (
	"encoding/json"
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func decodeSC14005(t *testing.T, client *connection.Client) *protobuf.SC_14005 {
	t.Helper()

	buffer := client.Buffer.Bytes()
	if len(buffer) == 0 {
		t.Fatalf("expected response buffer")
	}
	packetID := packets.GetPacketId(0, &buffer)
	if packetID != 14005 {
		t.Fatalf("expected packet 14005, got %d", packetID)
	}
	packetSize := packets.GetPacketSize(0, &buffer) + 2
	if len(buffer) < packetSize {
		t.Fatalf("expected packet size %d, got %d", packetSize, len(buffer))
	}
	payloadStart := packets.HEADER_SIZE
	payloadEnd := payloadStart + (packetSize - packets.HEADER_SIZE)
	var response protobuf.SC_14005
	if err := proto.Unmarshal(buffer[payloadStart:payloadEnd], &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	client.Buffer.Reset()
	return &response
}

func createTestEquipment(t *testing.T, id uint32, next uint32, transUseGold uint32, transUseItem json.RawMessage) {
	t.Helper()

	eq := orm.Equipment{
		ID:                id,
		DestroyGold:       0,
		EquipLimit:        0,
		Group:             0,
		Important:         0,
		Level:             1,
		Next:              int(next),
		Prev:              0,
		RestoreGold:       0,
		TransUseGold:      transUseGold,
		Type:              0,
		UpgradeFormulaID:  nil,
		DestroyItem:       nil,
		RestoreItem:       nil,
		ShipTypeForbidden: nil,
		TransUseItem:      transUseItem,
	}
	if err := orm.GormDB.Create(&eq).Error; err != nil {
		t.Fatalf("failed to create equipment %d: %v", id, err)
	}
}

func TestUpgradeEquipmentInBag14004EquipIDZero(t *testing.T) {
	client := &connection.Client{Commander: &orm.Commander{CommanderID: 9004701, AccountID: 9004701, Name: "equip-zero"}}

	payload := &protobuf.CS_14004{EquipId: proto.Uint32(0), Lv: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.UpgradeEquipmentInBag14004(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	response := decodeSC14005(t, client)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}

func TestUpgradeEquipmentInBag14004LvZero(t *testing.T) {
	client := &connection.Client{Commander: &orm.Commander{CommanderID: 9004702, AccountID: 9004702, Name: "lv-zero"}}

	payload := &protobuf.CS_14004{EquipId: proto.Uint32(100), Lv: proto.Uint32(0)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.UpgradeEquipmentInBag14004(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	response := decodeSC14005(t, client)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}

func TestUpgradeEquipmentInBag14004SuccessMutatesBag(t *testing.T) {
	commander := orm.Commander{CommanderID: 9004703, AccountID: 9004703, Name: "upgrade-success"}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}

	baseID := uint32(991000)
	upgradedID := uint32(991001)
	createTestEquipment(t, baseID, upgradedID, 0, nil)
	createTestEquipment(t, upgradedID, 0, 0, nil)

	if err := orm.GormDB.Create(&orm.OwnedEquipment{CommanderID: commander.CommanderID, EquipmentID: baseID, Count: 1}).Error; err != nil {
		t.Fatalf("failed to create owned equipment: %v", err)
	}

	client := &connection.Client{Commander: &commander}
	payload := &protobuf.CS_14004{EquipId: proto.Uint32(baseID), Lv: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.UpgradeEquipmentInBag14004(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	response := decodeSC14005(t, client)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}

	var base orm.OwnedEquipment
	if err := orm.GormDB.Where("commander_id = ? AND equipment_id = ?", commander.CommanderID, baseID).First(&base).Error; err == nil {
		t.Fatalf("expected base equipment to be removed")
	}

	var upgraded orm.OwnedEquipment
	if err := orm.GormDB.Where("commander_id = ? AND equipment_id = ?", commander.CommanderID, upgradedID).First(&upgraded).Error; err != nil {
		t.Fatalf("expected upgraded equipment to be added: %v", err)
	}
	if upgraded.Count != 1 {
		t.Fatalf("expected upgraded count 1, got %d", upgraded.Count)
	}
}

func TestUpgradeEquipmentInBag14004MissingChainDoesNotMutate(t *testing.T) {
	commander := orm.Commander{CommanderID: 9004704, AccountID: 9004704, Name: "upgrade-missing"}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}

	baseID := uint32(992000)
	createTestEquipment(t, baseID, 0, 0, nil)
	if err := orm.GormDB.Create(&orm.OwnedEquipment{CommanderID: commander.CommanderID, EquipmentID: baseID, Count: 1}).Error; err != nil {
		t.Fatalf("failed to create owned equipment: %v", err)
	}

	client := &connection.Client{Commander: &commander}
	payload := &protobuf.CS_14004{EquipId: proto.Uint32(baseID), Lv: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.UpgradeEquipmentInBag14004(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	response := decodeSC14005(t, client)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}

	var base orm.OwnedEquipment
	if err := orm.GormDB.Where("commander_id = ? AND equipment_id = ?", commander.CommanderID, baseID).First(&base).Error; err != nil {
		t.Fatalf("expected base equipment to remain: %v", err)
	}
	if base.Count != 1 {
		t.Fatalf("expected base count 1, got %d", base.Count)
	}
}

func TestUpgradeEquipmentInBag14004ChargesUpgradeCosts(t *testing.T) {
	commander := orm.Commander{CommanderID: 9004705, AccountID: 9004705, Name: "upgrade-costs"}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: commander.CommanderID, ResourceID: 1, Amount: 100}).Error; err != nil {
		t.Fatalf("failed to create gold resource: %v", err)
	}
	if err := orm.GormDB.Create(&orm.CommanderItem{CommanderID: commander.CommanderID, ItemID: 200, Count: 3}).Error; err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	baseID := uint32(993000)
	upgradedID := uint32(993001)
	createTestEquipment(t, baseID, upgradedID, 10, json.RawMessage(`[[200,1]]`))
	createTestEquipment(t, upgradedID, 0, 0, nil)
	if err := orm.GormDB.Create(&orm.OwnedEquipment{CommanderID: commander.CommanderID, EquipmentID: baseID, Count: 1}).Error; err != nil {
		t.Fatalf("failed to create owned equipment: %v", err)
	}

	client := &connection.Client{Commander: &commander}
	payload := &protobuf.CS_14004{EquipId: proto.Uint32(baseID), Lv: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.UpgradeEquipmentInBag14004(&buf, client); err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	response := decodeSC14005(t, client)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}

	var gold orm.OwnedResource
	if err := orm.GormDB.Where("commander_id = ? AND resource_id = ?", commander.CommanderID, 1).First(&gold).Error; err != nil {
		t.Fatalf("expected gold row: %v", err)
	}
	if gold.Amount != 90 {
		t.Fatalf("expected gold 90, got %d", gold.Amount)
	}

	var item orm.CommanderItem
	if err := orm.GormDB.Where("commander_id = ? AND item_id = ?", commander.CommanderID, 200).First(&item).Error; err != nil {
		t.Fatalf("expected item row: %v", err)
	}
	if item.Count != 2 {
		t.Fatalf("expected item count 2, got %d", item.Count)
	}
}
