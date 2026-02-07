package answer

import (
	"fmt"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestEventGiveUpSuccessClearsCollection(t *testing.T) {
	client := setupEventCollectionStartTest(t)

	seedEventCollectionTemplate(t, 101, `{"id":101,"collect_time":1800,"ship_num":2,"ship_lv":1,"ship_type":[],"oil":0,"drop_oil_max":0,"drop_gold_max":0,"over_time":0,"type":1,"max_team":0}`)
	seedEventCollectionShipTemplate(t, 1001, 1)
	seedEventCollectionShipTemplate(t, 1002, 1)
	ship1 := seedEventCollectionOwnedShip(t, client.Commander.CommanderID, 1001, 1)
	ship2 := seedEventCollectionOwnedShip(t, client.Commander.CommanderID, 1002, 1)
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	start := protobuf.CS_13003{Id: proto.Uint32(101), ShipIdList: []uint32{ship1.ID, ship2.ID}}
	startData, err := proto.Marshal(&start)
	if err != nil {
		t.Fatalf("marshal start: %v", err)
	}
	startBuf := startData
	if _, _, err := EventCollectionStart(&startBuf, client); err != nil {
		t.Fatalf("start: %v", err)
	}

	client.Buffer.Reset()
	giveUp := protobuf.CS_13007{Id: proto.Uint32(101)}
	data, err := proto.Marshal(&giveUp)
	if err != nil {
		t.Fatalf("marshal give up: %v", err)
	}
	buffer := data
	if _, _, err := EventGiveUp(&buffer, client); err != nil {
		t.Fatalf("give up: %v", err)
	}

	var response protobuf.SC_13008
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}

	var stored orm.EventCollection
	if err := orm.GormDB.First(&stored, "commander_id = ? AND collection_id = ?", client.Commander.CommanderID, 101).Error; err != nil {
		t.Fatalf("load stored: %v", err)
	}
	if stored.FinishTime != 0 || stored.StartTime != 0 {
		t.Fatalf("expected times cleared")
	}
	if len(stored.ShipIDs) != 0 {
		t.Fatalf("expected ship ids cleared")
	}
}

func TestEventGiveUpMissingCollectionFails(t *testing.T) {
	client := setupEventCollectionStartTest(t)
	client.Buffer.Reset()

	giveUp := protobuf.CS_13007{Id: proto.Uint32(999)}
	data, _ := proto.Marshal(&giveUp)
	buffer := data
	if _, _, err := EventGiveUp(&buffer, client); err != nil {
		t.Fatalf("give up: %v", err)
	}

	var response protobuf.SC_13008
	decodeResponse(t, client, &response)
	if response.GetResult() != 2 {
		t.Fatalf("expected result 2, got %d", response.GetResult())
	}
}

func TestEventGiveUpExpiredFinishTimeFails(t *testing.T) {
	client := setupEventCollectionStartTest(t)
	seedEventCollectionTemplate(t, 101, `{"id":101,"collect_time":1,"ship_num":1,"ship_lv":1,"ship_type":[],"oil":0,"drop_oil_max":0,"drop_gold_max":0,"over_time":0,"type":1,"max_team":0}`)
	if err := orm.GormDB.Create(&orm.EventCollection{CommanderID: client.Commander.CommanderID, CollectionID: 101, StartTime: 1, FinishTime: uint32(time.Now().Unix()) - 10, ShipIDs: orm.ToInt64List([]uint32{7})}).Error; err != nil {
		t.Fatalf("seed event: %v", err)
	}

	client.Buffer.Reset()
	giveUp := protobuf.CS_13007{Id: proto.Uint32(101)}
	data, _ := proto.Marshal(&giveUp)
	buffer := data
	if _, _, err := EventGiveUp(&buffer, client); err != nil {
		t.Fatalf("give up: %v", err)
	}

	var response protobuf.SC_13008
	decodeResponse(t, client, &response)
	if response.GetResult() != 2 {
		t.Fatalf("expected result 2, got %d", response.GetResult())
	}
}

func TestEventGiveUpOvertimeFails(t *testing.T) {
	client := setupEventCollectionStartTest(t)
	now := uint32(time.Now().Unix())
	overTime := now - 1
	seedEventCollectionTemplate(t, 101, fmt.Sprintf(`{"id":101,"collect_time":1800,"ship_num":1,"ship_lv":1,"ship_type":[],"oil":0,"drop_oil_max":0,"drop_gold_max":0,"over_time":%d,"type":1,"max_team":0}`, overTime))
	if err := orm.GormDB.Create(&orm.EventCollection{CommanderID: client.Commander.CommanderID, CollectionID: 101, StartTime: now - 5, FinishTime: now + 100, ShipIDs: orm.ToInt64List([]uint32{7})}).Error; err != nil {
		t.Fatalf("seed event: %v", err)
	}

	client.Buffer.Reset()
	giveUp := protobuf.CS_13007{Id: proto.Uint32(101)}
	data, _ := proto.Marshal(&giveUp)
	buffer := data
	if _, _, err := EventGiveUp(&buffer, client); err != nil {
		t.Fatalf("give up: %v", err)
	}

	var response protobuf.SC_13008
	decodeResponse(t, client, &response)
	if response.GetResult() != 3 {
		t.Fatalf("expected result 3, got %d", response.GetResult())
	}
}
