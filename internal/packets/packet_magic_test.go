package packets_test

import (
	"testing"

	"github.com/ggmolly/belfast/internal/packets"
)

func TestPacketParsing(t *testing.T) {
	buffer := []byte{0x00, 0x0A, 0x00, 0x04, 0xD2, 0x00, 0x02}
	if size := packets.GetPacketSize(0, &buffer); size != 10 {
		t.Fatalf("expected size 10, got %d", size)
	}
	if id := packets.GetPacketId(0, &buffer); id != 1234 {
		t.Fatalf("expected id 1234, got %d", id)
	}
	if index := packets.GetPacketIndex(0, &buffer); index != 2 {
		t.Fatalf("expected index 2, got %d", index)
	}

	prefixed := []byte{0xFF, 0xEE, 0xDD, 0x00, 0x0A, 0x00, 0x04, 0xD2, 0x00, 0x02}
	if size := packets.GetPacketSize(3, &prefixed); size != 10 {
		t.Fatalf("expected size 10, got %d", size)
	}
	if id := packets.GetPacketId(3, &prefixed); id != 1234 {
		t.Fatalf("expected id 1234, got %d", id)
	}
	if index := packets.GetPacketIndex(3, &prefixed); index != 2 {
		t.Fatalf("expected index 2, got %d", index)
	}
}

func FuzzPacketParsing(f *testing.F) {
	f.Add(uint8(0), []byte{0x00, 0x0A, 0x00, 0x04, 0xD2, 0x00, 0x02})
	f.Add(uint8(3), []byte{0xFF, 0xEE, 0xDD, 0x00, 0x0A, 0x00, 0x04, 0xD2, 0x00, 0x02})

	f.Fuzz(func(t *testing.T, offset uint8, buffer []byte) {
		offsetInt := int(offset)
		if offsetInt < 0 || len(buffer) < offsetInt+7 {
			return
		}

		expectedSize := int(buffer[0+offsetInt])<<8 + int(buffer[1+offsetInt])
		expectedID := int(buffer[3+offsetInt])<<8 + int(buffer[4+offsetInt])
		expectedIndex := int(buffer[5+offsetInt])<<8 + int(buffer[6+offsetInt])

		if size := packets.GetPacketSize(offsetInt, &buffer); size != expectedSize {
			t.Fatalf("expected size %d, got %d", expectedSize, size)
		}
		if id := packets.GetPacketId(offsetInt, &buffer); id != expectedID {
			t.Fatalf("expected id %d, got %d", expectedID, id)
		}
		if index := packets.GetPacketIndex(offsetInt, &buffer); index != expectedIndex {
			t.Fatalf("expected index %d, got %d", expectedIndex, index)
		}
	})
}
