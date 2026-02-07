package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/packets"
	"google.golang.org/protobuf/proto"
)

func decodePacketAt(t *testing.T, client *connection.Client, offset int, expectedID int, message proto.Message) int {
	t.Helper()
	buffer := client.Buffer.Bytes()
	if len(buffer) == 0 {
		t.Fatalf("expected response buffer")
	}
	packetId := packets.GetPacketId(offset, &buffer)
	if packetId != expectedID {
		t.Fatalf("expected packet %d, got %d", expectedID, packetId)
	}
	packetSize := packets.GetPacketSize(offset, &buffer) + 2
	if len(buffer) < offset+packetSize {
		t.Fatalf("expected packet size %d, got %d", offset+packetSize, len(buffer))
	}
	payloadStart := offset + packets.HEADER_SIZE
	payloadEnd := payloadStart + (packetSize - packets.HEADER_SIZE)
	if err := proto.Unmarshal(buffer[payloadStart:payloadEnd], message); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	return offset + packetSize
}
