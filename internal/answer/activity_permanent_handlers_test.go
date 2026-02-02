package answer

import (
	"strconv"
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func seedPermanentActivityConfig(t *testing.T, activityID uint32) {
	key := "6000"
	payload := `{"id":6000,"type":18,"time":"stop","config_data":[[35000]]}`
	if activityID != 6000 {
		key = strconv.FormatUint(uint64(activityID), 10)
		payload = `{"id":` + key + `,"type":18,"time":"stop","config_data":[[35000]]}`
	}
	seedConfigEntry(t, "ShareCfg/activity_task_permanent.json", key, `{"id":`+key+`}`)
	seedConfigEntry(t, "ShareCfg/activity_template.json", key, payload)
}

func decodePacketAtOffset(t *testing.T, buffer []byte, offset int, message proto.Message, expectedID int) int {
	packetID := packets.GetPacketId(offset, &buffer)
	if packetID != expectedID {
		t.Fatalf("expected packet %d, got %d", expectedID, packetID)
	}
	packetSize := packets.GetPacketSize(offset, &buffer) + 2
	payloadStart := offset + packets.HEADER_SIZE
	payloadEnd := offset + packetSize
	if err := proto.Unmarshal(buffer[payloadStart:payloadEnd], message); err != nil {
		t.Fatalf("unmarshal packet %d failed: %v", expectedID, err)
	}
	return offset + packetSize
}

func TestActivityPermanentStartSuccess(t *testing.T) {
	client := setupConfigTest(t)
	seedPermanentActivityConfig(t, 6000)

	payload := protobuf.CS_11206{ActivityId: proto.Uint32(6000)}
	data, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal request failed: %v", err)
	}
	buffer := data
	if _, _, err := ActivityPermanentStart(&buffer, client); err != nil {
		t.Fatalf("activity permanent start failed: %v", err)
	}

	state, err := orm.GetOrCreateActivityPermanentState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load permanent state failed: %v", err)
	}
	if state.CurrentActivityID != 6000 {
		t.Fatalf("expected current activity to be 6000")
	}

	responseBuffer := client.Buffer.Bytes()
	var update protobuf.SC_11201
	offset := decodePacketAtOffset(t, responseBuffer, 0, &update, 11201)
	if update.GetActivityInfo().GetId() != 6000 {
		t.Fatalf("expected activity info id 6000")
	}
	var response protobuf.SC_11207
	decodePacketAtOffset(t, responseBuffer, offset, &response, 11207)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}
}

func TestActivityPermanentStartRejectsFinished(t *testing.T) {
	client := setupConfigTest(t)
	seedPermanentActivityConfig(t, 6000)
	state := orm.ActivityPermanentState{
		CommanderID:         client.Commander.CommanderID,
		CurrentActivityID:   0,
		FinishedActivityIDs: orm.ToInt64List([]uint32{6000}),
	}
	if err := orm.GormDB.Create(&state).Error; err != nil {
		t.Fatalf("seed permanent state failed: %v", err)
	}

	payload := protobuf.CS_11206{ActivityId: proto.Uint32(6000)}
	data, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal request failed: %v", err)
	}
	buffer := data
	if _, _, err := ActivityPermanentStart(&buffer, client); err != nil {
		t.Fatalf("activity permanent start failed: %v", err)
	}

	responseBuffer := client.Buffer.Bytes()
	var response protobuf.SC_11207
	decodePacketAtOffset(t, responseBuffer, 0, &response, 11207)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}

func TestActivityPermanentFinishSuccess(t *testing.T) {
	client := setupConfigTest(t)
	seedPermanentActivityConfig(t, 6000)
	stateRecord := orm.ActivityPermanentState{
		CommanderID:         client.Commander.CommanderID,
		CurrentActivityID:   6000,
		FinishedActivityIDs: orm.ToInt64List([]uint32{}),
	}
	if err := orm.GormDB.Create(&stateRecord).Error; err != nil {
		t.Fatalf("seed permanent state failed: %v", err)
	}

	payload := protobuf.CS_11208{ActivityId: proto.Uint32(6000)}
	data, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal request failed: %v", err)
	}
	buffer := data
	if _, _, err := ActivityPermanentFinish(&buffer, client); err != nil {
		t.Fatalf("activity permanent finish failed: %v", err)
	}

	state, err := orm.GetOrCreateActivityPermanentState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load permanent state failed: %v", err)
	}
	if state.CurrentActivityID != 0 {
		t.Fatalf("expected current activity to be cleared")
	}
	if !state.HasFinished(6000) {
		t.Fatalf("expected finished list to include 6000")
	}

	responseBuffer := client.Buffer.Bytes()
	var response protobuf.SC_11209
	decodePacketAtOffset(t, responseBuffer, 0, &response, 11209)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0")
	}
}

func TestActivityPermanentFinishRejectsNotCurrent(t *testing.T) {
	client := setupConfigTest(t)
	seedPermanentActivityConfig(t, 6000)

	payload := protobuf.CS_11208{ActivityId: proto.Uint32(6000)}
	data, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal request failed: %v", err)
	}
	buffer := data
	if _, _, err := ActivityPermanentFinish(&buffer, client); err != nil {
		t.Fatalf("activity permanent finish failed: %v", err)
	}

	responseBuffer := client.Buffer.Bytes()
	var response protobuf.SC_11209
	decodePacketAtOffset(t, responseBuffer, 0, &response, 11209)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}
