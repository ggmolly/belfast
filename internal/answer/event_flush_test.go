package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestEventFlushReturnsCollections(t *testing.T) {
	client := setupEventCollectionStartTest(t)
	seedEventCollectionTemplate(t, 101, `{"id":101,"collect_time":1800,"ship_num":1,"ship_lv":1,"ship_type":[],"oil":0,"drop_oil_max":0,"drop_gold_max":0,"over_time":1234,"type":1,"max_team":0}`)
	seedEventCollectionTemplate(t, 102, `{"id":102,"collect_time":1800,"ship_num":1,"ship_lv":1,"ship_type":[],"oil":0,"drop_oil_max":0,"drop_gold_max":0,"over_time":0,"type":1,"max_team":0}`)

	clearTable(t, &orm.EventCollection{})
	if err := orm.GormDB.Create(&orm.EventCollection{CommanderID: client.Commander.CommanderID, CollectionID: 102, StartTime: 1, FinishTime: 20, ShipIDs: orm.ToInt64List([]uint32{9})}).Error; err != nil {
		t.Fatalf("seed event: %v", err)
	}
	if err := orm.GormDB.Create(&orm.EventCollection{CommanderID: client.Commander.CommanderID, CollectionID: 101, StartTime: 2, FinishTime: 10, ShipIDs: orm.ToInt64List([]uint32{7, 8})}).Error; err != nil {
		t.Fatalf("seed event: %v", err)
	}

	request := protobuf.CS_13009{Type: proto.Uint32(0)}
	data, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	buffer := data
	if _, _, err := EventFlush(&buffer, client); err != nil {
		t.Fatalf("event flush: %v", err)
	}

	respBuf := client.Buffer.Bytes()
	var response protobuf.SC_13010
	decodePacketAtOffset(t, respBuf, 0, &response, 13010)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}
	list := response.GetCollectionList()
	if len(list) != 2 {
		t.Fatalf("expected 2 collections, got %d", len(list))
	}
	if list[0].GetId() != 101 || list[0].GetFinishTime() != 10 || list[0].GetOverTime() != 1234 {
		t.Fatalf("unexpected first collection: %+v", list[0])
	}
	if len(list[0].GetShipIdList()) != 2 || list[0].GetShipIdList()[0] != 7 || list[0].GetShipIdList()[1] != 8 {
		t.Fatalf("unexpected ship ids: %v", list[0].GetShipIdList())
	}
	if list[1].GetId() != 102 || list[1].GetFinishTime() != 20 || list[1].GetOverTime() != 0 {
		t.Fatalf("unexpected second collection: %+v", list[1])
	}
}

func TestEventFlushEmptyList(t *testing.T) {
	client := setupEventCollectionStartTest(t)
	clearTable(t, &orm.EventCollection{})

	request := protobuf.CS_13009{Type: proto.Uint32(0)}
	data, _ := proto.Marshal(&request)
	buffer := data
	if _, _, err := EventFlush(&buffer, client); err != nil {
		t.Fatalf("event flush: %v", err)
	}

	respBuf := client.Buffer.Bytes()
	var response protobuf.SC_13010
	_ = decodePacketAtOffset(t, respBuf, 0, &response, 13010)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}
	if len(response.GetCollectionList()) != 0 {
		t.Fatalf("expected empty collection list")
	}
}

func TestEventFlushInvalidPayload(t *testing.T) {
	client := setupEventCollectionStartTest(t)
	buffer := []byte{0x01, 0x02, 0x03}
	_, packetID, err := EventFlush(&buffer, client)
	if err == nil {
		t.Fatalf("expected error")
	}
	if packetID != 13010 {
		t.Fatalf("expected packet id 13010, got %d", packetID)
	}
	if client.Buffer.Len() != 0 {
		t.Fatalf("expected no response written")
	}
}
