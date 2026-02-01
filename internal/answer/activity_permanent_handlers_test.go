package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestActivityPermanentStart(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.ConfigEntry{})
	seedConfigEntry(t, "ShareCfg/activity_task_permanent.json", "6000", `{"id":6000}`)
	seedConfigEntry(t, "ShareCfg/activity_template.json", "6000", `{"id":6000,"type":18,"time":"stop"}`)

	payload := protobuf.CS_11206{ActivityId: proto.Uint32(6000)}
	data, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload failed: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := ActivityPermanentStart(&data, client); err != nil {
		t.Fatalf("activity permanent start failed: %v", err)
	}

	var response protobuf.SC_11207
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}
	state, err := orm.GetOrCreatePermanentActivityState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load permanent state failed: %v", err)
	}
	if state.PermanentNow != 6000 {
		t.Fatalf("expected permanent now to be 6000")
	}
}

func TestActivityPermanentFinish(t *testing.T) {
	client := setupHandlerCommander(t)
	clearTable(t, &orm.ConfigEntry{})
	seedConfigEntry(t, "ShareCfg/activity_task_permanent.json", "6000", `{"id":6000}`)
	seedConfigEntry(t, "ShareCfg/activity_template.json", "6000", `{"id":6000,"type":18,"time":"stop"}`)
	entry := orm.PermanentActivityState{CommanderID: client.Commander.CommanderID, PermanentNow: 6000}
	if err := orm.GormDB.Create(&entry).Error; err != nil {
		t.Fatalf("seed permanent activity state failed: %v", err)
	}

	payload := protobuf.CS_11208{ActivityId: proto.Uint32(6000)}
	data, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload failed: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := ActivityPermanentFinish(&data, client); err != nil {
		t.Fatalf("activity permanent finish failed: %v", err)
	}

	var response protobuf.SC_11209
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}
	fetchedState, err := orm.GetOrCreatePermanentActivityState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load permanent state failed: %v", err)
	}
	if fetchedState.PermanentNow != 0 {
		t.Fatalf("expected permanent now to be 0")
	}
	finished := orm.ToUint32List(fetchedState.FinishedActivityIDs)
	if len(finished) != 1 || finished[0] != 6000 {
		t.Fatalf("expected finished list to include 6000")
	}
}
