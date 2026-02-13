package answer_test

import (
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
	execAnswerExternalTestSQLT(t, "DELETE FROM owned_equipments")
	execAnswerExternalTestSQLT(t, "DELETE FROM equipments")
	execAnswerExternalTestSQLT(t, "DELETE FROM commander_items")
	execAnswerExternalTestSQLT(t, "DELETE FROM owned_resources")
	execAnswerExternalTestSQLT(t, "DELETE FROM commanders")
	if err := orm.CreateCommanderRoot(1001, 1001, "Destroy Equipments Tester", 0, 0); err != nil {
		t.Fatalf("create commander: %v", err)
	}
	commander := orm.Commander{CommanderID: 1001}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func TestDestroyEquipmentsSuccessAwardsRewards(t *testing.T) {
	client := setupDestroyEquipmentsTest(t)
	execAnswerExternalTestSQLT(t, "INSERT INTO equipments (id, destroy_gold, destroy_item, ship_type_forbidden) VALUES ($1, $2, $3::jsonb, $4::jsonb)", int64(2001), int64(10), `[[300,1],[301,2]]`, `[]`)
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_equipments (commander_id, equipment_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(2001), int64(5))
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1), int64(50))
	execAnswerExternalTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(300), int64(7))
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
	execAnswerExternalTestSQLT(t, "INSERT INTO equipments (id, destroy_gold, destroy_item, ship_type_forbidden) VALUES ($1, $2, $3::jsonb, $4::jsonb)", int64(2002), int64(9), `[[302,1]]`, `[]`)
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_equipments (commander_id, equipment_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(2002), int64(1))
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1), int64(5))
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
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_equipments (commander_id, equipment_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(9001), int64(2))
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
	execAnswerExternalTestSQLT(t, "INSERT INTO equipments (id, destroy_gold, destroy_item, ship_type_forbidden) VALUES ($1, $2, $3::jsonb, $4::jsonb)", int64(2100), int64(2), `[[400,1]]`, `[]`)
	execAnswerExternalTestSQLT(t, "INSERT INTO equipments (id, destroy_gold, destroy_item, ship_type_forbidden) VALUES ($1, $2, $3::jsonb, $4::jsonb)", int64(2101), int64(3), `[[401,2]]`, `[]`)
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_equipments (commander_id, equipment_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(2100), int64(4))
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_equipments (commander_id, equipment_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(2101), int64(1))
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1), int64(0))
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
	execAnswerExternalTestSQLT(t, "INSERT INTO equipments (id, destroy_gold, destroy_item, ship_type_forbidden) VALUES ($1, $2, $3::jsonb, $4::jsonb)", int64(2200), int64(1), `[[500,1]]`, `[]`)
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_equipments (commander_id, equipment_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(2200), int64(1))
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1), int64(0))
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
