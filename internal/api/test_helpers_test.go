package api_test

import (
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/packets"
	"google.golang.org/protobuf/proto"
)

func decodeTestPacket(t *testing.T, buffer []byte, expectedId int, message proto.Message) int {
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
	return packetId
}

func seedDb() {
	tx := orm.GormDB.Begin()
	resources := []orm.Resource{
		{ID: 1, Name: "Gold"},
		{ID: 2, Name: "Fake resource"},
	}
	items := []orm.Item{
		{ID: 20001, Name: "Wisdom Cube"},
		{ID: 45, Name: "Fake Item"},
		{ID: 60, Name: "Fake Item 2"},
	}
	for _, r := range resources {
		tx.Save(&r)
	}
	for _, i := range items {
		tx.Save(&i)
	}
	if err := tx.Commit().Error; err != nil {
		panic(err)
	}
}
