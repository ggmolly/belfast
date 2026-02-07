package answer_test

import (
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestEquipSpWeaponSuccess(t *testing.T) {
	client := setupSpWeaponClient(t)
	clearTable(t, &orm.OwnedShip{})

	ship := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1}
	if err := orm.GormDB.Create(&ship).Error; err != nil {
		t.Fatalf("failed to create ship: %v", err)
	}
	spweapon := orm.OwnedSpWeapon{OwnerID: client.Commander.CommanderID, TemplateID: 1001}
	if err := orm.GormDB.Create(&spweapon).Error; err != nil {
		t.Fatalf("failed to create spweapon: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("failed to reload commander: %v", err)
	}

	payload := &protobuf.CS_14201{ShipId: proto.Uint32(ship.ID), SpweaponId: proto.Uint32(spweapon.ID)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.EquipSpWeapon(&buf, client); err != nil {
		t.Fatalf("EquipSpWeapon failed: %v", err)
	}

	response := &protobuf.SC_14202{}
	decodeTestPacket(t, client, 14202, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}

	var stored orm.OwnedSpWeapon
	if err := orm.GormDB.First(&stored, "owner_id = ? AND id = ?", client.Commander.CommanderID, spweapon.ID).Error; err != nil {
		t.Fatalf("failed to load spweapon: %v", err)
	}
	if stored.EquippedShipID != ship.ID {
		t.Fatalf("expected spweapon to be equipped to ship %d, got %d", ship.ID, stored.EquippedShipID)
	}
}

func TestEquipSpWeaponInvalidShipNoPersist(t *testing.T) {
	client := setupSpWeaponClient(t)
	clearTable(t, &orm.OwnedShip{})

	spweapon := orm.OwnedSpWeapon{OwnerID: client.Commander.CommanderID, TemplateID: 1001, EquippedShipID: 0}
	if err := orm.GormDB.Create(&spweapon).Error; err != nil {
		t.Fatalf("failed to create spweapon: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("failed to reload commander: %v", err)
	}

	payload := &protobuf.CS_14201{ShipId: proto.Uint32(999999), SpweaponId: proto.Uint32(spweapon.ID)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.EquipSpWeapon(&buf, client); err != nil {
		t.Fatalf("EquipSpWeapon failed: %v", err)
	}

	response := &protobuf.SC_14202{}
	decodeTestPacket(t, client, 14202, response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}

	var stored orm.OwnedSpWeapon
	if err := orm.GormDB.First(&stored, "owner_id = ? AND id = ?", client.Commander.CommanderID, spweapon.ID).Error; err != nil {
		t.Fatalf("failed to load spweapon: %v", err)
	}
	if stored.EquippedShipID != 0 {
		t.Fatalf("expected spweapon equip state unchanged, got %d", stored.EquippedShipID)
	}
}

func TestEquipSpWeaponMovesAndUnequipsOthersOnTargetShip(t *testing.T) {
	client := setupSpWeaponClient(t)
	clearTable(t, &orm.OwnedShip{})

	shipA := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1}
	shipB := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 2}
	if err := orm.GormDB.Create(&shipA).Error; err != nil {
		t.Fatalf("failed to create ship A: %v", err)
	}
	if err := orm.GormDB.Create(&shipB).Error; err != nil {
		t.Fatalf("failed to create ship B: %v", err)
	}

	spweapon := orm.OwnedSpWeapon{OwnerID: client.Commander.CommanderID, TemplateID: 1001, EquippedShipID: shipA.ID}
	otherOnB := orm.OwnedSpWeapon{OwnerID: client.Commander.CommanderID, TemplateID: 1002, EquippedShipID: shipB.ID}
	if err := orm.GormDB.Create(&spweapon).Error; err != nil {
		t.Fatalf("failed to create spweapon: %v", err)
	}
	if err := orm.GormDB.Create(&otherOnB).Error; err != nil {
		t.Fatalf("failed to create other spweapon: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("failed to reload commander: %v", err)
	}

	payload := &protobuf.CS_14201{ShipId: proto.Uint32(shipB.ID), SpweaponId: proto.Uint32(spweapon.ID)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.EquipSpWeapon(&buf, client); err != nil {
		t.Fatalf("EquipSpWeapon failed: %v", err)
	}

	response := &protobuf.SC_14202{}
	decodeTestPacket(t, client, 14202, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}

	var moved orm.OwnedSpWeapon
	if err := orm.GormDB.First(&moved, "owner_id = ? AND id = ?", client.Commander.CommanderID, spweapon.ID).Error; err != nil {
		t.Fatalf("failed to load moved spweapon: %v", err)
	}
	if moved.EquippedShipID != shipB.ID {
		t.Fatalf("expected spweapon to be equipped to ship %d, got %d", shipB.ID, moved.EquippedShipID)
	}

	var unequipped orm.OwnedSpWeapon
	if err := orm.GormDB.First(&unequipped, "owner_id = ? AND id = ?", client.Commander.CommanderID, otherOnB.ID).Error; err != nil {
		t.Fatalf("failed to load other spweapon: %v", err)
	}
	if unequipped.EquippedShipID != 0 {
		t.Fatalf("expected other spweapon to be unequipped, got %d", unequipped.EquippedShipID)
	}
}
