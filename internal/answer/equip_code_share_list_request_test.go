package answer_test

import (
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestEquipCodeShareListRequestReturnsEmptyLists(t *testing.T) {
	client := &connection.Client{}
	payload := &protobuf.CS_17601{Shipgroup: proto.Uint32(1001)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	if _, _, err := answer.EquipCodeShareListRequest(&buf, client); err != nil {
		t.Fatalf("EquipCodeShareListRequest failed: %v", err)
	}

	response := &protobuf.SC_17602{}
	packetId := decodeTestPacket(t, client, 17602, response)
	if packetId != 17602 {
		t.Fatalf("expected packet 17602, got %d", packetId)
	}
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if len(response.GetInfos()) != 0 {
		t.Fatalf("expected infos to be empty, got %d", len(response.GetInfos()))
	}
	if len(response.GetRecentInfos()) != 0 {
		t.Fatalf("expected recent_infos to be empty, got %d", len(response.GetRecentInfos()))
	}
}
