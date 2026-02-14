package answer_test

import (
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func seedSpWeaponShipTemplate(t *testing.T, templateID uint32) {
	t.Helper()
	execAnswerExternalTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (template_id) DO NOTHING", int64(templateID), "SpWeapon Ship", "SpWeapon Ship", int64(1), int64(1), int64(1), int64(1), int64(0))
}

func TestEquipSpWeaponSuccess(t *testing.T) {
	client := setupSpWeaponClient(t)
	clearTable(t, &orm.OwnedShip{})
	seedSpWeaponShipTemplate(t, 1001)

	ship := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1001}
	if err := ship.Create(); err != nil {
		t.Fatalf("failed to create ship: %v", err)
	}
	created, err := orm.CreateOwnedSpWeapon(client.Commander.CommanderID, 1001)
	if err != nil {
		t.Fatalf("failed to create spweapon: %v", err)
	}
	spweapon := *created
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

	stored, err := orm.GetOwnedSpWeapon(client.Commander.CommanderID, spweapon.ID)
	if err != nil {
		t.Fatalf("failed to load spweapon: %v", err)
	}
	if stored.EquippedShipID != ship.ID {
		t.Fatalf("expected spweapon to be equipped to ship %d, got %d", ship.ID, stored.EquippedShipID)
	}
}

func TestEquipSpWeaponInvalidShipNoPersist(t *testing.T) {
	client := setupSpWeaponClient(t)
	clearTable(t, &orm.OwnedShip{})
	seedSpWeaponShipTemplate(t, 1001)

	created, err := orm.CreateOwnedSpWeapon(client.Commander.CommanderID, 1001)
	if err != nil {
		t.Fatalf("failed to create spweapon: %v", err)
	}
	spweapon := *created
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

	stored, err := orm.GetOwnedSpWeapon(client.Commander.CommanderID, spweapon.ID)
	if err != nil {
		t.Fatalf("failed to load spweapon: %v", err)
	}
	if stored.EquippedShipID != 0 {
		t.Fatalf("expected spweapon equip state unchanged, got %d", stored.EquippedShipID)
	}
}

func TestEquipSpWeaponMovesAndUnequipsOthersOnTargetShip(t *testing.T) {
	client := setupSpWeaponClient(t)
	clearTable(t, &orm.OwnedShip{})
	seedSpWeaponShipTemplate(t, 1001)
	seedSpWeaponShipTemplate(t, 1002)

	shipA := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1001}
	shipB := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1002}
	if err := shipA.Create(); err != nil {
		t.Fatalf("failed to create ship A: %v", err)
	}
	if err := shipB.Create(); err != nil {
		t.Fatalf("failed to create ship B: %v", err)
	}

	spweapon := orm.OwnedSpWeapon{OwnerID: client.Commander.CommanderID, TemplateID: 1001, EquippedShipID: shipA.ID}
	otherOnB := orm.OwnedSpWeapon{OwnerID: client.Commander.CommanderID, TemplateID: 1002, EquippedShipID: shipB.ID}
	created, err := orm.CreateOwnedSpWeapon(client.Commander.CommanderID, 1001)
	if err != nil {
		t.Fatalf("failed to create spweapon: %v", err)
	}
	spweapon.ID = created.ID
	if err := orm.SaveOwnedSpWeapon(&spweapon); err != nil {
		t.Fatalf("failed to update spweapon: %v", err)
	}
	createdOther, err := orm.CreateOwnedSpWeapon(client.Commander.CommanderID, 1002)
	if err != nil {
		t.Fatalf("failed to create other spweapon: %v", err)
	}
	otherOnB.ID = createdOther.ID
	if err := orm.SaveOwnedSpWeapon(&otherOnB); err != nil {
		t.Fatalf("failed to update other spweapon: %v", err)
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

	moved, err := orm.GetOwnedSpWeapon(client.Commander.CommanderID, spweapon.ID)
	if err != nil {
		t.Fatalf("failed to load moved spweapon: %v", err)
	}
	if moved.EquippedShipID != shipB.ID {
		t.Fatalf("expected spweapon to be equipped to ship %d, got %d", shipB.ID, moved.EquippedShipID)
	}

	unequipped, err := orm.GetOwnedSpWeapon(client.Commander.CommanderID, otherOnB.ID)
	if err != nil {
		t.Fatalf("failed to load other spweapon: %v", err)
	}
	if unequipped.EquippedShipID != 0 {
		t.Fatalf("expected other spweapon to be unequipped, got %d", unequipped.EquippedShipID)
	}
}
