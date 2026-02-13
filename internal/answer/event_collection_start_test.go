package answer

import (
	"fmt"
	"testing"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func setupEventCollectionStartTest(t *testing.T) *connection.Client {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.EventCollection{})
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.OwnedShip{})
	clearTable(t, &orm.Ship{})
	return client
}

func seedEventCollectionTemplate(t *testing.T, collectionID uint32, payload string) {
	t.Helper()
	seedConfigEntry(t, "ShareCfg/collection_template.json", fmt.Sprintf("%d", collectionID), payload)
}

func seedEventCollectionShipTemplate(t *testing.T, templateID uint32, shipType uint32) {
	t.Helper()
	execAnswerTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(templateID), "Test", "Test", int64(2), int64(1), int64(shipType), int64(1), int64(0))
}

func seedEventCollectionOwnedShip(t *testing.T, commanderID uint32, templateID uint32, level uint32) orm.OwnedShip {
	t.Helper()
	owned := orm.OwnedShip{OwnerID: commanderID, ShipID: templateID, Level: level}
	if err := owned.Create(); err != nil {
		t.Fatalf("seed owned ship: %v", err)
	}
	return owned
}

func seedEventCollectionResource(t *testing.T, commanderID uint32, resourceID uint32, amount uint32) {
	t.Helper()
	execAnswerTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(commanderID), int64(resourceID), int64(amount))
}

func TestEventCollectionStartSuccess(t *testing.T) {
	client := setupEventCollectionStartTest(t)

	seedEventCollectionTemplate(t, 101, `{"id":101,"collect_time":1800,"ship_num":2,"ship_lv":10,"ship_type":[1,2],"oil":5,"drop_oil_max":0,"drop_gold_max":0,"over_time":0,"type":1,"max_team":1}`)
	seedEventCollectionShipTemplate(t, 1001, 1)
	seedEventCollectionShipTemplate(t, 1002, 2)
	ship1 := seedEventCollectionOwnedShip(t, client.Commander.CommanderID, 1001, 10)
	ship2 := seedEventCollectionOwnedShip(t, client.Commander.CommanderID, 1002, 1)
	seedEventCollectionResource(t, client.Commander.CommanderID, 2, 10)
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	request := protobuf.CS_13003{Id: proto.Uint32(101), ShipIdList: []uint32{ship1.ID, ship2.ID}}
	data, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	buffer := data
	if _, _, err := EventCollectionStart(&buffer, client); err != nil {
		t.Fatalf("event collection start: %v", err)
	}

	respBuf := client.Buffer.Bytes()
	var update protobuf.SC_13011
	offset := decodePacketAtOffset(t, respBuf, 0, &update, 13011)
	if len(update.GetCollection()) != 1 || update.GetCollection()[0].GetId() != 101 {
		t.Fatalf("expected update for collection 101")
	}
	var response protobuf.SC_13004
	decodePacketAtOffset(t, respBuf, offset, &response, 13004)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}

	stored, err := orm.GetEventCollection(nil, client.Commander.CommanderID, 101)
	if err != nil {
		t.Fatalf("expected event persisted: %v", err)
	}
	if stored.FinishTime == 0 || len(stored.ShipIDs) != 2 {
		t.Fatalf("expected finish time and 2 ship ids")
	}
	oil := queryAnswerTestInt64(t, "SELECT amount FROM owned_resources WHERE commander_id = $1 AND resource_id = $2", int64(client.Commander.CommanderID), int64(2))
	if oil != 5 {
		t.Fatalf("expected oil 5, got %d", oil)
	}
}

func TestEventCollectionStartInsufficientOil(t *testing.T) {
	client := setupEventCollectionStartTest(t)
	seedEventCollectionTemplate(t, 101, `{"id":101,"collect_time":1800,"ship_num":2,"ship_lv":1,"ship_type":[1],"oil":10,"drop_oil_max":0,"drop_gold_max":0,"over_time":0,"type":1,"max_team":0}`)
	seedEventCollectionShipTemplate(t, 1001, 1)
	seedEventCollectionShipTemplate(t, 1002, 1)
	ship1 := seedEventCollectionOwnedShip(t, client.Commander.CommanderID, 1001, 1)
	ship2 := seedEventCollectionOwnedShip(t, client.Commander.CommanderID, 1002, 1)
	seedEventCollectionResource(t, client.Commander.CommanderID, 2, 5)
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	request := protobuf.CS_13003{Id: proto.Uint32(101), ShipIdList: []uint32{ship1.ID, ship2.ID}}
	data, _ := proto.Marshal(&request)
	buffer := data
	if _, _, err := EventCollectionStart(&buffer, client); err != nil {
		t.Fatalf("handler: %v", err)
	}
	var response protobuf.SC_13004
	decodeResponse(t, client, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}

func TestEventCollectionStartRejectsShipsNotOwned(t *testing.T) {
	client := setupEventCollectionStartTest(t)
	seedEventCollectionTemplate(t, 101, `{"id":101,"collect_time":1800,"ship_num":2,"ship_lv":1,"ship_type":[],"oil":0,"drop_oil_max":0,"drop_gold_max":0,"over_time":0,"type":1,"max_team":0}`)
	seedEventCollectionShipTemplate(t, 1001, 1)
	ship1 := seedEventCollectionOwnedShip(t, client.Commander.CommanderID, 1001, 1)
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	request := protobuf.CS_13003{Id: proto.Uint32(101), ShipIdList: []uint32{ship1.ID, 999999}}
	data, _ := proto.Marshal(&request)
	buffer := data
	if _, _, err := EventCollectionStart(&buffer, client); err != nil {
		t.Fatalf("handler: %v", err)
	}
	var response protobuf.SC_13004
	decodeResponse(t, client, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}

func TestEventCollectionStartRejectsShipCountMismatch(t *testing.T) {
	client := setupEventCollectionStartTest(t)
	seedEventCollectionTemplate(t, 101, `{"id":101,"collect_time":1800,"ship_num":2,"ship_lv":1,"ship_type":[],"oil":0,"drop_oil_max":0,"drop_gold_max":0,"over_time":0,"type":1,"max_team":0}`)
	seedEventCollectionShipTemplate(t, 1001, 1)
	ship1 := seedEventCollectionOwnedShip(t, client.Commander.CommanderID, 1001, 1)
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	request := protobuf.CS_13003{Id: proto.Uint32(101), ShipIdList: []uint32{ship1.ID}}
	data, _ := proto.Marshal(&request)
	buffer := data
	if _, _, err := EventCollectionStart(&buffer, client); err != nil {
		t.Fatalf("handler: %v", err)
	}
	var response protobuf.SC_13004
	decodeResponse(t, client, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}

func TestEventCollectionStartRejectsShipLevelTooLow(t *testing.T) {
	client := setupEventCollectionStartTest(t)
	seedEventCollectionTemplate(t, 101, `{"id":101,"collect_time":1800,"ship_num":2,"ship_lv":10,"ship_type":[],"oil":0,"drop_oil_max":0,"drop_gold_max":0,"over_time":0,"type":1,"max_team":0}`)
	seedEventCollectionShipTemplate(t, 1001, 1)
	seedEventCollectionShipTemplate(t, 1002, 1)
	ship1 := seedEventCollectionOwnedShip(t, client.Commander.CommanderID, 1001, 1)
	ship2 := seedEventCollectionOwnedShip(t, client.Commander.CommanderID, 1002, 1)
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	request := protobuf.CS_13003{Id: proto.Uint32(101), ShipIdList: []uint32{ship1.ID, ship2.ID}}
	data, _ := proto.Marshal(&request)
	buffer := data
	if _, _, err := EventCollectionStart(&buffer, client); err != nil {
		t.Fatalf("handler: %v", err)
	}
	var response protobuf.SC_13004
	decodeResponse(t, client, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}

func TestEventCollectionStartRejectsShipTypeNotAllowed(t *testing.T) {
	client := setupEventCollectionStartTest(t)
	seedEventCollectionTemplate(t, 101, `{"id":101,"collect_time":1800,"ship_num":2,"ship_lv":1,"ship_type":[1],"oil":0,"drop_oil_max":0,"drop_gold_max":0,"over_time":0,"type":1,"max_team":0}`)
	seedEventCollectionShipTemplate(t, 1001, 1)
	seedEventCollectionShipTemplate(t, 1002, 2)
	ship1 := seedEventCollectionOwnedShip(t, client.Commander.CommanderID, 1001, 1)
	ship2 := seedEventCollectionOwnedShip(t, client.Commander.CommanderID, 1002, 1)
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	request := protobuf.CS_13003{Id: proto.Uint32(101), ShipIdList: []uint32{ship1.ID, ship2.ID}}
	data, _ := proto.Marshal(&request)
	buffer := data
	if _, _, err := EventCollectionStart(&buffer, client); err != nil {
		t.Fatalf("handler: %v", err)
	}
	var response protobuf.SC_13004
	decodeResponse(t, client, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}

func TestEventCollectionStartRejectsMaxTeamExceeded(t *testing.T) {
	client := setupEventCollectionStartTest(t)
	seedEventCollectionTemplate(t, 101, `{"id":101,"collect_time":1800,"ship_num":2,"ship_lv":1,"ship_type":[],"oil":0,"drop_oil_max":0,"drop_gold_max":0,"over_time":0,"type":1,"max_team":1}`)
	seedEventCollectionShipTemplate(t, 1001, 1)
	seedEventCollectionShipTemplate(t, 1002, 1)
	ship1 := seedEventCollectionOwnedShip(t, client.Commander.CommanderID, 1001, 1)
	ship2 := seedEventCollectionOwnedShip(t, client.Commander.CommanderID, 1002, 1)
	entry, err := orm.GetOrCreateActiveEvent(nil, client.Commander.CommanderID, 999)
	if err != nil {
		t.Fatalf("seed existing event: %v", err)
	}
	entry.StartTime = 1
	entry.FinishTime = 2
	entry.ShipIDs = orm.ToInt64List([]uint32{ship1.ID})
	if err := orm.SaveEventCollection(nil, entry); err != nil {
		t.Fatalf("seed existing event: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	request := protobuf.CS_13003{Id: proto.Uint32(101), ShipIdList: []uint32{ship1.ID, ship2.ID}}
	data, _ := proto.Marshal(&request)
	buffer := data
	if _, _, err := EventCollectionStart(&buffer, client); err != nil {
		t.Fatalf("handler: %v", err)
	}
	var response protobuf.SC_13004
	decodeResponse(t, client, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}

func TestEventCollectionStartRejectsDuplicateEvent(t *testing.T) {
	client := setupEventCollectionStartTest(t)
	seedEventCollectionTemplate(t, 101, `{"id":101,"collect_time":1800,"ship_num":2,"ship_lv":1,"ship_type":[],"oil":0,"drop_oil_max":0,"drop_gold_max":0,"over_time":0,"type":1,"max_team":0}`)
	seedEventCollectionShipTemplate(t, 1001, 1)
	seedEventCollectionShipTemplate(t, 1002, 1)
	ship1 := seedEventCollectionOwnedShip(t, client.Commander.CommanderID, 1001, 1)
	ship2 := seedEventCollectionOwnedShip(t, client.Commander.CommanderID, 1002, 1)
	entry, err := orm.GetOrCreateActiveEvent(nil, client.Commander.CommanderID, 101)
	if err != nil {
		t.Fatalf("seed existing event: %v", err)
	}
	entry.StartTime = 1
	entry.FinishTime = 2
	entry.ShipIDs = orm.ToInt64List([]uint32{ship1.ID, ship2.ID})
	if err := orm.SaveEventCollection(nil, entry); err != nil {
		t.Fatalf("seed existing event: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	request := protobuf.CS_13003{Id: proto.Uint32(101), ShipIdList: []uint32{ship1.ID, ship2.ID}}
	data, _ := proto.Marshal(&request)
	buffer := data
	if _, _, err := EventCollectionStart(&buffer, client); err != nil {
		t.Fatalf("handler: %v", err)
	}
	var response protobuf.SC_13004
	decodeResponse(t, client, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}

func TestEventCollectionStartRejectsExpiredEvent(t *testing.T) {
	client := setupEventCollectionStartTest(t)
	seedEventCollectionTemplate(t, 101, `{"id":101,"collect_time":1800,"ship_num":1,"ship_lv":1,"ship_type":[],"oil":0,"drop_oil_max":0,"drop_gold_max":0,"over_time":1,"type":1,"max_team":0}`)
	seedEventCollectionShipTemplate(t, 1001, 1)
	ship1 := seedEventCollectionOwnedShip(t, client.Commander.CommanderID, 1001, 1)
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	request := protobuf.CS_13003{Id: proto.Uint32(101), ShipIdList: []uint32{ship1.ID}}
	data, _ := proto.Marshal(&request)
	buffer := data
	if _, _, err := EventCollectionStart(&buffer, client); err != nil {
		t.Fatalf("handler: %v", err)
	}
	var response protobuf.SC_13004
	decodeResponse(t, client, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}

func TestEventCollectionStartRejectsResourceCapExceeded(t *testing.T) {
	client := setupEventCollectionStartTest(t)
	seedEventCollectionTemplate(t, 101, `{"id":101,"collect_time":1800,"ship_num":1,"ship_lv":1,"ship_type":[],"oil":0,"drop_oil_max":0,"drop_gold_max":100,"over_time":0,"type":1,"max_team":0}`)
	seedEventCollectionShipTemplate(t, 1001, 1)
	ship1 := seedEventCollectionOwnedShip(t, client.Commander.CommanderID, 1001, 1)
	seedEventCollectionResource(t, client.Commander.CommanderID, 1, 100)
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	request := protobuf.CS_13003{Id: proto.Uint32(101), ShipIdList: []uint32{ship1.ID}}
	data, _ := proto.Marshal(&request)
	buffer := data
	if _, _, err := EventCollectionStart(&buffer, client); err != nil {
		t.Fatalf("handler: %v", err)
	}
	var response protobuf.SC_13004
	decodeResponse(t, client, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}
