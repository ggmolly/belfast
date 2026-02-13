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
	execAnswerExternalTestSQLT(t, "DELETE FROM owned_ship_equipments")
	execAnswerExternalTestSQLT(t, "DELETE FROM owned_ships")
	execAnswerExternalTestSQLT(t, "DELETE FROM owned_resources")
	execAnswerExternalTestSQLT(t, "DELETE FROM commander_items")
	execAnswerExternalTestSQLT(t, "DELETE FROM commander_misc_items")
	execAnswerExternalTestSQLT(t, "DELETE FROM equipments")
	execAnswerExternalTestSQLT(t, "DELETE FROM ships")
	execAnswerExternalTestSQLT(t, "DELETE FROM commanders")

	if err := orm.CreateCommanderRoot(commanderID, commanderID, "Upgrade Equip Ship", 0, 0); err != nil {
		t.Fatalf("create commander: %v", err)
	}
	commander := orm.Commander{CommanderID: commanderID}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func TestUpgradeEquipmentOnShip14002SuccessUpdatesSlotAndChargesCosts(t *testing.T) {
	client := setupUpgradeEquipmentOnShip14002Test(t, 9014002)

	shipTemplate := orm.Ship{TemplateID: 1001, Name: "Ship", EnglishName: "Ship", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	execAnswerExternalTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(shipTemplate.TemplateID), shipTemplate.Name, shipTemplate.EnglishName, int64(shipTemplate.RarityID), int64(shipTemplate.Star), int64(shipTemplate.Type), int64(shipTemplate.Nationality), int64(shipTemplate.BuildTime))
	ownedShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1001, ID: 50001, Level: 1, MaxLevel: 50}
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, level, max_level, create_time, change_name_timestamp) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())", int64(ownedShip.ID), int64(ownedShip.OwnerID), int64(ownedShip.ShipID), int64(ownedShip.Level), int64(ownedShip.MaxLevel))

	execAnswerExternalTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1), int64(100))
	execAnswerExternalTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(200), int64(3))

	baseID := uint32(995000)
	upgradedID := uint32(995001)
	createTestEquipment(t, baseID, upgradedID, 10, json.RawMessage(`[[200,1]]`))
	createTestEquipment(t, upgradedID, 0, 0, nil)

	execAnswerExternalTestSQLT(t, "INSERT INTO owned_ship_equipments (owner_id, ship_id, pos, equip_id, skin_id) VALUES ($1, $2, $3, $4, $5)", int64(client.Commander.CommanderID), int64(ownedShip.ID), int64(1), int64(baseID), int64(123))

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

	updated, err := orm.GetOwnedShipEquipment(client.Commander.CommanderID, ownedShip.ID, 1)
	if err != nil {
		t.Fatalf("load ship equipment: %v", err)
	}
	if updated.EquipID != upgradedID {
		t.Fatalf("expected equip id %d, got %d", upgradedID, updated.EquipID)
	}
	if updated.SkinID != 123 {
		t.Fatalf("expected skin id 123 to remain, got %d", updated.SkinID)
	}

	gold := queryAnswerExternalTestInt64(t, "SELECT amount FROM owned_resources WHERE commander_id = $1 AND resource_id = $2", int64(client.Commander.CommanderID), int64(1))
	if gold != 90 {
		t.Fatalf("expected gold 90, got %d", gold)
	}

	item := queryAnswerExternalTestInt64(t, "SELECT count FROM commander_items WHERE commander_id = $1 AND item_id = $2", int64(client.Commander.CommanderID), int64(200))
	if item != 2 {
		t.Fatalf("expected item count 2, got %d", item)
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
	execAnswerExternalTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(shipTemplate.TemplateID), shipTemplate.Name, shipTemplate.EnglishName, int64(shipTemplate.RarityID), int64(shipTemplate.Star), int64(shipTemplate.Type), int64(shipTemplate.Nationality), int64(shipTemplate.BuildTime))
	ownedShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1001, ID: 50002, Level: 1, MaxLevel: 50}
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, level, max_level, create_time, change_name_timestamp) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())", int64(ownedShip.ID), int64(ownedShip.OwnerID), int64(ownedShip.ShipID), int64(ownedShip.Level), int64(ownedShip.MaxLevel))

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
	execAnswerExternalTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(shipTemplate.TemplateID), shipTemplate.Name, shipTemplate.EnglishName, int64(shipTemplate.RarityID), int64(shipTemplate.Star), int64(shipTemplate.Type), int64(shipTemplate.Nationality), int64(shipTemplate.BuildTime))
	ownedShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1001, ID: 50003, Level: 1, MaxLevel: 50}
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, level, max_level, create_time, change_name_timestamp) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())", int64(ownedShip.ID), int64(ownedShip.OwnerID), int64(ownedShip.ShipID), int64(ownedShip.Level), int64(ownedShip.MaxLevel))

	execAnswerExternalTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1), int64(5))
	execAnswerExternalTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(200), int64(3))

	baseID := uint32(996000)
	upgradedID := uint32(996001)
	createTestEquipment(t, baseID, upgradedID, 10, json.RawMessage(`[[200,1]]`))
	createTestEquipment(t, upgradedID, 0, 0, nil)
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_ship_equipments (owner_id, ship_id, pos, equip_id, skin_id) VALUES ($1, $2, $3, $4, $5)", int64(client.Commander.CommanderID), int64(ownedShip.ID), int64(1), int64(baseID), int64(0))

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

	entry, err := orm.GetOwnedShipEquipment(client.Commander.CommanderID, ownedShip.ID, 1)
	if err != nil {
		t.Fatalf("load ship equipment: %v", err)
	}
	if entry.EquipID != baseID {
		t.Fatalf("expected equip id %d, got %d", baseID, entry.EquipID)
	}
}

func TestUpgradeEquipmentOnShip14002NotEnoughItemsDoesNotMutate(t *testing.T) {
	client := setupUpgradeEquipmentOnShip14002Test(t, 9014006)

	shipTemplate := orm.Ship{TemplateID: 1001, Name: "Ship", EnglishName: "Ship", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	execAnswerExternalTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(shipTemplate.TemplateID), shipTemplate.Name, shipTemplate.EnglishName, int64(shipTemplate.RarityID), int64(shipTemplate.Star), int64(shipTemplate.Type), int64(shipTemplate.Nationality), int64(shipTemplate.BuildTime))
	ownedShip := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1001, ID: 50004, Level: 1, MaxLevel: 50}
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, level, max_level, create_time, change_name_timestamp) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())", int64(ownedShip.ID), int64(ownedShip.OwnerID), int64(ownedShip.ShipID), int64(ownedShip.Level), int64(ownedShip.MaxLevel))

	execAnswerExternalTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1), int64(100))
	execAnswerExternalTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(200), int64(0))

	baseID := uint32(997000)
	upgradedID := uint32(997001)
	createTestEquipment(t, baseID, upgradedID, 10, json.RawMessage(`[[200,1]]`))
	createTestEquipment(t, upgradedID, 0, 0, nil)
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_ship_equipments (owner_id, ship_id, pos, equip_id, skin_id) VALUES ($1, $2, $3, $4, $5)", int64(client.Commander.CommanderID), int64(ownedShip.ID), int64(1), int64(baseID), int64(0))

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

	entry, err := orm.GetOwnedShipEquipment(client.Commander.CommanderID, ownedShip.ID, 1)
	if err != nil {
		t.Fatalf("load ship equipment: %v", err)
	}
	if entry.EquipID != baseID {
		t.Fatalf("expected equip id %d, got %d", baseID, entry.EquipID)
	}
}
