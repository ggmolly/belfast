package tests

import (
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

// TODO: Update these tests if CS_10991 or CS_11720 handling changes.

func decodeTestPacket(t *testing.T, client *connection.Client, expectedId int, message proto.Message) int {
	buffer := client.Buffer.Bytes()
	if len(buffer) == 0 {
		t.Fatalf("expected response buffer")
	}
	packetId := packets.GetPacketId(0, &buffer)
	if packetId != expectedId {
		t.Fatalf("expected packet %d, got %d", expectedId, packetId)
	}
	packetSize := packets.GetPacketSize(0, &buffer) + 2
	if len(buffer) < packetSize {
		t.Fatalf("expected packet size %d, got %d", packetSize, len(buffer))
	}
	payloadStart := packets.HEADER_SIZE
	payloadEnd := payloadStart + (packetSize - packets.HEADER_SIZE)
	if err := proto.Unmarshal(buffer[payloadStart:payloadEnd], message); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	client.Buffer.Reset()
	return packetId
}

func TestGameTrackingAck(t *testing.T) {
	client := &connection.Client{}
	payload := &protobuf.CS_10991{}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.GameTracking(&buf, client); err != nil {
		t.Fatalf("GameTracking failed: %v", err)
	}
	response := &protobuf.CS_10992{}
	packetId := decodeTestPacket(t, client, 10992, response)
	if packetId != 10992 {
		t.Fatalf("expected packet 10992, got %d", packetId)
	}
}

func TestJuustagramReadTipAck(t *testing.T) {
	client := &connection.Client{}
	payload := &protobuf.CS_11720{ChatGroupIdList: []uint32{1}}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.JuustagramReadTip(&buf, client); err != nil {
		t.Fatalf("JuustagramReadTip failed: %v", err)
	}
	response := &protobuf.SC_11721{}
	packetId := decodeTestPacket(t, client, 11721, response)
	if packetId != 11721 {
		t.Fatalf("expected packet 11721, got %d", packetId)
	}
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
}
