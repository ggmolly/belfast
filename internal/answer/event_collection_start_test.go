package answer

import (
	"fmt"
	"testing"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/packets"
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

func seedEventCollectionOwnedShip(t *testing.T, client *connection.Client, templateID uint32, level uint32) orm.OwnedShip {
	t.Helper()
	owned := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: templateID, Level: level}
	if err := owned.Create(); err != nil {
		t.Fatalf("seed owned ship: %v", err)
	}
	owned.Level = level
	if err := owned.Update(); err != nil {
		t.Fatalf("seed owned ship update: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}
	return owned
}

func seedEventCollectionResource(t *testing.T, client *connection.Client, resourceID uint32, amount uint32) {
	t.Helper()
	if err := client.Commander.SetResource(resourceID, amount); err != nil {
		t.Fatalf("seed resource: %v", err)
	}
}

func TestEventCollectionStartSuccess(t *testing.T) {
	client := setupEventCollectionStartTest(t)

	seedEventCollectionTemplate(t, 101, `{"id":101,"collect_time":1800,"ship_num":2,"ship_lv":10,"ship_type":[1,2],"oil":5,"drop_oil_max":0,"drop_gold_max":0,"over_time":0,"type":1,"max_team":1}`)
	seedEventCollectionShipTemplate(t, 1001, 1)
	seedEventCollectionShipTemplate(t, 1002, 2)
	ship1 := seedEventCollectionOwnedShip(t, client, 1001, 10)
	ship2 := seedEventCollectionOwnedShip(t, client, 1002, 10)
	seedEventCollectionResource(t, client, 2, 10)
	if client.Commander.GetResourceCount(2) != 10 {
		t.Fatalf("expected oil 10 in commander map, got %d", client.Commander.GetResourceCount(2))
	}
	owned1 := client.Commander.OwnedShipsMap[ship1.ID]
	owned2 := client.Commander.OwnedShipsMap[ship2.ID]
	if owned1 == nil || owned2 == nil {
		t.Fatalf("expected seeded ships in commander map")
	}
	if owned1.Level < 10 || owned2.Level < 10 {
		t.Fatalf("expected seeded ship levels >= 10, got %d and %d", owned1.Level, owned2.Level)
	}
	if owned1.Ship.Type == 0 || owned2.Ship.Type == 0 {
		t.Fatalf("expected seeded ship types, got %d and %d", owned1.Ship.Type, owned2.Ship.Type)
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
	if len(respBuf) == 0 {
		t.Fatalf("expected response packets")
	}
	first := packets.GetPacketId(0, &respBuf)
	var response protobuf.SC_13004
	seenUpdate := false
	offset := 0
	if first == 13004 {
		offset = decodePacketAtOffset(t, respBuf, 0, &response, 13004)
	} else if first == 13011 {
		var firstUpdate protobuf.SC_13011
		offset = decodePacketAtOffset(t, respBuf, 0, &firstUpdate, 13011)
		if len(firstUpdate.GetCollection()) == 1 && firstUpdate.GetCollection()[0].GetId() == 101 {
			seenUpdate = true
		}
		decodePacketAtOffset(t, respBuf, offset, &response, 13004)
	} else {
		t.Fatalf("unexpected first packet %d", first)
	}
	if !seenUpdate && offset < len(respBuf) {
		var nextUpdate protobuf.SC_13011
		decodePacketAtOffset(t, respBuf, offset, &nextUpdate, 13011)
		if len(nextUpdate.GetCollection()) == 1 && nextUpdate.GetCollection()[0].GetId() == 101 {
			seenUpdate = true
		}
	}
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if !seenUpdate {
		t.Fatalf("expected update for collection 101")
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
	ship1 := seedEventCollectionOwnedShip(t, client, 1001, 1)
	ship2 := seedEventCollectionOwnedShip(t, client, 1002, 1)
	seedEventCollectionResource(t, client, 2, 5)
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
	ship1 := seedEventCollectionOwnedShip(t, client, 1001, 1)
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
	ship1 := seedEventCollectionOwnedShip(t, client, 1001, 1)
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
	ship1 := seedEventCollectionOwnedShip(t, client, 1001, 1)
	ship2 := seedEventCollectionOwnedShip(t, client, 1002, 1)
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
	ship1 := seedEventCollectionOwnedShip(t, client, 1001, 1)
	ship2 := seedEventCollectionOwnedShip(t, client, 1002, 1)
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
	ship1 := seedEventCollectionOwnedShip(t, client, 1001, 1)
	ship2 := seedEventCollectionOwnedShip(t, client, 1002, 1)
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
	ship1 := seedEventCollectionOwnedShip(t, client, 1001, 1)
	ship2 := seedEventCollectionOwnedShip(t, client, 1002, 1)
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
	ship1 := seedEventCollectionOwnedShip(t, client, 1001, 1)
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
	ship1 := seedEventCollectionOwnedShip(t, client, 1001, 1)
	seedEventCollectionResource(t, client, 1, 100)
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
