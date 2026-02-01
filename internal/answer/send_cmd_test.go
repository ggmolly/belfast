package answer_test

import (
	"strings"
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestSendCmdIntoReturnsOk(t *testing.T) {
	client := &connection.Client{}
	payload := &protobuf.CS_11100{Cmd: proto.String("into")}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.SendCmd(&buf, client); err != nil {
		t.Fatalf("SendCmd failed: %v", err)
	}

	response := &protobuf.SC_11101{}
	packetId := decodeTestPacket(t, client, 11101, response)
	if packetId != 11101 {
		t.Fatalf("expected packet 11101, got %d", packetId)
	}
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if !strings.Contains(response.GetMsg(), "CMD:into") || !strings.Contains(response.GetMsg(), "Result:ok") {
		t.Fatalf("expected ok message, got %q", response.GetMsg())
	}
}

func TestSendCmdWorldResetReturnsOk(t *testing.T) {
	client := &connection.Client{}
	payload := &protobuf.CS_11100{Cmd: proto.String("world"), Arg1: proto.String("reset")}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.SendCmd(&buf, client); err != nil {
		t.Fatalf("SendCmd failed: %v", err)
	}

	response := &protobuf.SC_11101{}
	packetId := decodeTestPacket(t, client, 11101, response)
	if packetId != 11101 {
		t.Fatalf("expected packet 11101, got %d", packetId)
	}
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if !strings.Contains(response.GetMsg(), "CMD:world") || !strings.Contains(response.GetMsg(), "Result:ok") {
		t.Fatalf("expected ok message, got %q", response.GetMsg())
	}
}

func TestSendCmdUnknownReturnsFail(t *testing.T) {
	client := &connection.Client{}
	payload := &protobuf.CS_11100{Cmd: proto.String("nope")}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.SendCmd(&buf, client); err != nil {
		t.Fatalf("SendCmd failed: %v", err)
	}

	response := &protobuf.SC_11101{}
	packetId := decodeTestPacket(t, client, 11101, response)
	if packetId != 11101 {
		t.Fatalf("expected packet 11101, got %d", packetId)
	}
	if response.GetResult() != 1 {
		t.Fatalf("expected result 1, got %d", response.GetResult())
	}
	if !strings.Contains(response.GetMsg(), "CMD:nope") || !strings.Contains(response.GetMsg(), "Result:fail") {
		t.Fatalf("expected fail message, got %q", response.GetMsg())
	}
}

func TestSendCmdKickDisconnects(t *testing.T) {
	client := &connection.Client{}
	payload := &protobuf.CS_11100{Cmd: proto.String("kick")}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.SendCmd(&buf, client); err != nil {
		t.Fatalf("SendCmd failed: %v", err)
	}

	buffer := client.Buffer.Bytes()
	if len(buffer) == 0 {
		t.Fatalf("expected response buffer")
	}
	packetId := packets.GetPacketId(0, &buffer)
	if packetId != 11101 {
		t.Fatalf("expected packet 11101, got %d", packetId)
	}
	packetSize := packets.GetPacketSize(0, &buffer) + 2
	if len(buffer) < packetSize {
		t.Fatalf("expected packet size %d, got %d", packetSize, len(buffer))
	}
	payloadStart := packets.HEADER_SIZE
	payloadEnd := payloadStart + (packetSize - packets.HEADER_SIZE)
	response := &protobuf.SC_11101{}
	if err := proto.Unmarshal(buffer[payloadStart:payloadEnd], response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}

	remaining := buffer[packetSize:]
	if len(remaining) == 0 {
		t.Fatalf("expected disconnect packet after kick response")
	}
	secondId := packets.GetPacketId(0, &remaining)
	if secondId != 10999 {
		t.Fatalf("expected packet 10999, got %d", secondId)
	}
	secondSize := packets.GetPacketSize(0, &remaining) + 2
	if len(remaining) < secondSize {
		t.Fatalf("expected disconnect packet size %d, got %d", secondSize, len(remaining))
	}
	secondPayloadStart := packets.HEADER_SIZE
	secondPayloadEnd := secondPayloadStart + (secondSize - packets.HEADER_SIZE)
	disconnect := &protobuf.SC_10999{}
	if err := proto.Unmarshal(remaining[secondPayloadStart:secondPayloadEnd], disconnect); err != nil {
		t.Fatalf("failed to unmarshal disconnect: %v", err)
	}
	if disconnect.GetReason() != uint32(consts.DR_CONNECTION_TO_SERVER_LOST) {
		t.Fatalf("expected disconnect reason %d, got %d", consts.DR_CONNECTION_TO_SERVER_LOST, disconnect.GetReason())
	}
}
