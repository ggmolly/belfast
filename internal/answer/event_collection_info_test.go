package answer

import (
	"fmt"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestEventCollectionInfoReturnsStoredEvents(t *testing.T) {
	client := setupEventCollectionStartTest(t)

	overTime := uint32(1234)
	seedEventCollectionTemplate(t, 101, fmt.Sprintf(`{"id":101,"collect_time":1800,"ship_num":1,"ship_lv":1,"ship_type":[],"oil":0,"drop_oil_max":0,"drop_gold_max":0,"over_time":%d,"type":1,"max_team":0}`, overTime))
	entry, err := orm.GetOrCreateActiveEvent(nil, client.Commander.CommanderID, 101)
	if err != nil {
		t.Fatalf("seed event: %v", err)
	}
	entry.StartTime = 1
	entry.FinishTime = uint32(time.Now().Unix()) + 100
	entry.ShipIDs = orm.ToInt64List([]uint32{7, 8})
	if err := orm.SaveEventCollection(nil, entry); err != nil {
		t.Fatalf("seed event: %v", err)
	}

	client.Buffer.Reset()
	buffer := []byte{}
	if _, _, err := EventCollectionInfo(&buffer, client); err != nil {
		t.Fatalf("event collection info: %v", err)
	}

	var response protobuf.SC_13002
	decodeResponse(t, client, &response)
	if response.GetMaxTeam() != 0 {
		t.Fatalf("expected max team 0")
	}
	if len(response.GetCollectionList()) != 1 {
		t.Fatalf("expected 1 collection, got %d", len(response.GetCollectionList()))
	}
	info := response.GetCollectionList()[0]
	if info.GetId() != 101 {
		t.Fatalf("expected collection id 101")
	}
	if info.GetFinishTime() == 0 {
		t.Fatalf("expected finish time set")
	}
	if info.GetOverTime() != overTime {
		t.Fatalf("expected overtime %d, got %d", overTime, info.GetOverTime())
	}
	if len(info.GetShipIdList()) != 2 || info.GetShipIdList()[0] != 7 || info.GetShipIdList()[1] != 8 {
		t.Fatalf("expected ship list [7 8]")
	}
}

func TestEventCollectionInfoSkipsInactiveRows(t *testing.T) {
	client := setupEventCollectionStartTest(t)
	seedEventCollectionTemplate(t, 101, `{"id":101,"collect_time":1800,"ship_num":1,"ship_lv":1,"ship_type":[],"oil":0,"drop_oil_max":0,"drop_gold_max":0,"over_time":0,"type":1,"max_team":0}`)
	entry, err := orm.GetOrCreateActiveEvent(nil, client.Commander.CommanderID, 101)
	if err != nil {
		t.Fatalf("seed event: %v", err)
	}
	entry.StartTime = 0
	entry.FinishTime = 0
	entry.ShipIDs = orm.Int64List{}
	if err := orm.SaveEventCollection(nil, entry); err != nil {
		t.Fatalf("seed event: %v", err)
	}

	client.Buffer.Reset()
	buffer := []byte{}
	if _, _, err := EventCollectionInfo(&buffer, client); err != nil {
		t.Fatalf("event collection info: %v", err)
	}

	var response protobuf.SC_13002
	decodeResponse(t, client, &response)
	if len(response.GetCollectionList()) != 0 {
		t.Fatalf("expected inactive collection to be skipped")
	}
}

func TestEventCollectionInfoOrdersByCollectionID(t *testing.T) {
	client := setupEventCollectionStartTest(t)
	seedEventCollectionTemplate(t, 102, `{"id":102,"collect_time":1,"ship_num":1,"ship_lv":1,"ship_type":[],"oil":0,"drop_oil_max":0,"drop_gold_max":0,"over_time":0,"type":1,"max_team":0}`)
	seedEventCollectionTemplate(t, 101, `{"id":101,"collect_time":1,"ship_num":1,"ship_lv":1,"ship_type":[],"oil":0,"drop_oil_max":0,"drop_gold_max":0,"over_time":0,"type":1,"max_team":0}`)

	now := uint32(time.Now().Unix())
	entry102, err := orm.GetOrCreateActiveEvent(nil, client.Commander.CommanderID, 102)
	if err != nil {
		t.Fatalf("seed event: %v", err)
	}
	entry102.StartTime = 1
	entry102.FinishTime = now + 100
	entry102.ShipIDs = orm.ToInt64List([]uint32{1})
	if err := orm.SaveEventCollection(nil, entry102); err != nil {
		t.Fatalf("seed event: %v", err)
	}
	entry101, err := orm.GetOrCreateActiveEvent(nil, client.Commander.CommanderID, 101)
	if err != nil {
		t.Fatalf("seed event: %v", err)
	}
	entry101.StartTime = 1
	entry101.FinishTime = now + 100
	entry101.ShipIDs = orm.ToInt64List([]uint32{2})
	if err := orm.SaveEventCollection(nil, entry101); err != nil {
		t.Fatalf("seed event: %v", err)
	}

	client.Buffer.Reset()
	buffer := []byte{}
	if _, _, err := EventCollectionInfo(&buffer, client); err != nil {
		t.Fatalf("event collection info: %v", err)
	}

	var response protobuf.SC_13002
	decodeResponse(t, client, &response)
	if len(response.GetCollectionList()) != 2 {
		t.Fatalf("expected 2 collections")
	}
	if response.GetCollectionList()[0].GetId() != 101 || response.GetCollectionList()[1].GetId() != 102 {
		t.Fatalf("expected collections ordered by id")
	}
}

func TestEventCollectionInfoWorksWithoutTemplate(t *testing.T) {
	client := setupEventCollectionStartTest(t)
	now := uint32(time.Now().Unix())
	entry, err := orm.GetOrCreateActiveEvent(nil, client.Commander.CommanderID, 101)
	if err != nil {
		t.Fatalf("seed event: %v", err)
	}
	entry.StartTime = 1
	entry.FinishTime = now + 100
	entry.ShipIDs = orm.ToInt64List([]uint32{7})
	if err := orm.SaveEventCollection(nil, entry); err != nil {
		t.Fatalf("seed event: %v", err)
	}

	client.Buffer.Reset()
	buffer := []byte{}
	if _, _, err := EventCollectionInfo(&buffer, client); err != nil {
		t.Fatalf("event collection info: %v", err)
	}

	var response protobuf.SC_13002
	decodeResponse(t, client, &response)
	if len(response.GetCollectionList()) != 1 {
		t.Fatalf("expected 1 collection")
	}
	if response.GetCollectionList()[0].GetOverTime() != 0 {
		t.Fatalf("expected overtime 0 when template missing")
	}
}

func TestEventCollectionInfoRoundTripsProto(t *testing.T) {
	client := setupEventCollectionStartTest(t)
	client.Buffer.Reset()
	buffer := []byte{}
	if _, _, err := EventCollectionInfo(&buffer, client); err != nil {
		t.Fatalf("event collection info: %v", err)
	}

	var response protobuf.SC_13002
	decodeResponse(t, client, &response)
	if response.MaxTeam == nil {
		t.Fatalf("expected required max_team field to be set")
	}
	// Ensure proto required fields are present.
	if _, err := proto.Marshal(&response); err != nil {
		t.Fatalf("marshal response: %v", err)
	}
}
