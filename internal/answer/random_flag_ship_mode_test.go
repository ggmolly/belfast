package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func decodeClientResponse(t *testing.T, client *connection.Client, response proto.Message) {
	t.Helper()
	data := client.Buffer.Bytes()
	if len(data) < 7 {
		t.Fatalf("expected buffer to include header and payload")
	}
	if err := proto.Unmarshal(data[7:], response); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
}

func TestChangeRandomFlagShipModeValid(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	payload := protobuf.CS_12206{Flag: proto.Uint32(2)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := ChangeRandomFlagShipMode(&buffer, client); err != nil {
		t.Fatalf("change random flagship mode failed: %v", err)
	}
	var response protobuf.SC_12207
	decodeClientResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	var commander orm.Commander
	if err := orm.GormDB.First(&commander, client.Commander.CommanderID).Error; err != nil {
		t.Fatalf("load commander: %v", err)
	}
	if commander.RandomShipMode != 2 {
		t.Fatalf("expected random ship mode 2, got %d", commander.RandomShipMode)
	}
	var flag orm.CommanderCommonFlag
	if err := orm.GormDB.First(&flag, "commander_id = ? AND flag_id = ?", commander.CommanderID, consts.RandomFlagShipMode).Error; err != nil {
		t.Fatalf("expected common flag to be set")
	}
}

func TestChangeRandomFlagShipModeInvalid(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	payload := protobuf.CS_12206{Flag: proto.Uint32(0)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := ChangeRandomFlagShipMode(&buffer, client); err != nil {
		t.Fatalf("change random flagship mode failed: %v", err)
	}
	var response protobuf.SC_12207
	decodeClientResponse(t, client, &response)
	if response.GetResult() != 1 {
		t.Fatalf("expected result 1, got %d", response.GetResult())
	}
	var commander orm.Commander
	if err := orm.GormDB.First(&commander, client.Commander.CommanderID).Error; err != nil {
		t.Fatalf("load commander: %v", err)
	}
	if commander.RandomShipMode != 0 {
		t.Fatalf("expected random ship mode 0, got %d", commander.RandomShipMode)
	}
}

func TestToggleRandomFlagShip(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	payload := protobuf.CS_12204{Flag: proto.Uint32(1)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := ToggleRandomFlagShip(&buffer, client); err != nil {
		t.Fatalf("toggle random flagship failed: %v", err)
	}
	var response protobuf.SC_12205
	decodeClientResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	var commander orm.Commander
	if err := orm.GormDB.First(&commander, client.Commander.CommanderID).Error; err != nil {
		t.Fatalf("load commander: %v", err)
	}
	if !commander.RandomFlagShipEnabled {
		t.Fatalf("expected random flagship enabled")
	}

	invalidPayload := protobuf.CS_12204{Flag: proto.Uint32(2)}
	invalidBuffer, err := proto.Marshal(&invalidPayload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := ToggleRandomFlagShip(&invalidBuffer, client); err != nil {
		t.Fatalf("toggle random flagship failed: %v", err)
	}
	var invalidResponse protobuf.SC_12205
	decodeClientResponse(t, client, &invalidResponse)
	if invalidResponse.GetResult() != 1 {
		t.Fatalf("expected result 1, got %d", invalidResponse.GetResult())
	}
}
