package answer_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func setupRevertEquipmentTest(t *testing.T) *connection.Client {
	t.Helper()
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	execAnswerExternalTestSQLT(t, "DELETE FROM owned_equipments")
	execAnswerExternalTestSQLT(t, "DELETE FROM equipments")
	execAnswerExternalTestSQLT(t, "DELETE FROM commander_items")
	execAnswerExternalTestSQLT(t, "DELETE FROM commander_misc_items")
	execAnswerExternalTestSQLT(t, "DELETE FROM owned_resources")
	execAnswerExternalTestSQLT(t, "DELETE FROM config_entries")
	execAnswerExternalTestSQLT(t, "DELETE FROM commanders")
	if err := orm.CreateCommanderRoot(901, 901, "Revert Equipment Tester", 0, 0); err != nil {
		t.Fatalf("create commander: %v", err)
	}
	seedRevertEquipmentItemDefs(t)
	commander := orm.Commander{CommanderID: 901}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func seedRevertEquipmentItemDefs(t *testing.T) {
	t.Helper()
	for _, itemID := range []uint32{15007, 200, 201} {
		execAnswerExternalTestSQLT(t, "INSERT INTO items (id, name, rarity, shop_id, type, virtual_type) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (id) DO NOTHING", int64(itemID), "Revert Item", int64(1), int64(0), int64(1), int64(0))
	}
}

func seedRevertEquipmentChain(t *testing.T) {
	t.Helper()
	seedConfigEntryRevertEquipment(t, 500, `{"id":500,"prev":0,"level":1,"trans_use_gold":10,"trans_use_item":[[200,1]],"ship_type_forbidden":[]}`)
	seedConfigEntryRevertEquipment(t, 501, `{"id":501,"prev":500,"level":2,"trans_use_gold":20,"trans_use_item":[[200,2],[201,1]],"ship_type_forbidden":[]}`)
	seedConfigEntryRevertEquipment(t, 502, `{"id":502,"prev":501,"level":3,"trans_use_gold":30,"trans_use_item":[[200,3]],"ship_type_forbidden":[]}`)
}

func seedConfigEntryRevertEquipment(t *testing.T, equipID uint32, payload string) {
	t.Helper()
	execAnswerExternalTestSQLT(t, "INSERT INTO config_entries (category, key, data) VALUES ($1, $2, $3::jsonb)", "sharecfgdata/equip_data_statistics.json", fmt.Sprintf("%d", equipID), payload)
}

func loadOwnedEquipmentCount(t *testing.T, commanderID uint32, equipmentID uint32) uint32 {
	t.Helper()
	value := queryAnswerExternalTestInt64(t, "SELECT COALESCE((SELECT count FROM owned_equipments WHERE commander_id = $1 AND equipment_id = $2), 0)", int64(commanderID), int64(equipmentID))
	return uint32(value)
}

func loadItemCount(t *testing.T, commanderID uint32, itemID uint32) uint32 {
	t.Helper()
	value := queryAnswerExternalTestInt64(t, "SELECT COALESCE((SELECT count FROM commander_items WHERE commander_id = $1 AND item_id = $2), 0)", int64(commanderID), int64(itemID))
	return uint32(value)
}

func loadResourceCount(t *testing.T, commanderID uint32, resourceID uint32) uint32 {
	t.Helper()
	value := queryAnswerExternalTestInt64(t, "SELECT COALESCE((SELECT amount FROM owned_resources WHERE commander_id = $1 AND resource_id = $2), 0)", int64(commanderID), int64(resourceID))
	return uint32(value)
}

func TestRevertEquipmentSuccess(t *testing.T) {
	client := setupRevertEquipmentTest(t)
	seedRevertEquipmentChain(t)
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_equipments (commander_id, equipment_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(502), int64(1))
	execAnswerExternalTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(15007), int64(1))
	execAnswerExternalTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(200), int64(10))
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1), int64(100))
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
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_equipments (commander_id, equipment_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(502), int64(1))
	execAnswerExternalTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(200), int64(10))
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1), int64(100))
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
	execAnswerExternalTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(15007), int64(1))
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(1), int64(100))
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
	execAnswerExternalTestSQLT(t, "INSERT INTO equipments (id, prev, level, ship_type_forbidden) VALUES ($1, $2, $3, $4::jsonb)", int64(600), int64(0), int64(1), `[]`)
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_equipments (commander_id, equipment_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(600), int64(1))
	execAnswerExternalTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(15007), int64(1))
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
	execAnswerExternalTestSQLT(t, "INSERT INTO owned_equipments (commander_id, equipment_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(700), int64(1))
	execAnswerExternalTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(15007), int64(1))
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
	execAnswerExternalTestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(15007), int64(1))
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
