package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestGetPhantomQuestProgress(t *testing.T) {
	client := setupHandlerCommander(t)
	request := protobuf.CS_12212{ShipIdList: []uint32{1, 2, 3}}
	data, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	client.Buffer.Reset()
	_, packetID, err := GetPhantomQuestProgress(&data, client)
	if err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	if packetID != 12213 {
		t.Fatalf("expected packet id 12213, got %d", packetID)
	}
	buf := client.Buffer.Bytes()
	if got := packets.GetPacketId(0, &buf); got != 12213 {
		t.Fatalf("expected response packet 12213, got %d", got)
	}

	var response protobuf.SC_12213
	decodeResponse(t, client, &response)
	entries := response.GetShipCountList()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	for i, shipID := range []uint32{1, 2, 3} {
		entry := entries[i]
		if entry == nil {
			t.Fatalf("expected entry %d", i)
		}
		if entry.GetKey() != shipID {
			t.Fatalf("expected key %d, got %d", shipID, entry.GetKey())
		}
		if entry.GetValue() != 0 {
			t.Fatalf("expected value 0, got %d", entry.GetValue())
		}
	}
}

func TestGetPhantomQuestProgressEmptyRequest(t *testing.T) {
	client := setupHandlerCommander(t)
	request := protobuf.CS_12212{}
	data, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	client.Buffer.Reset()
	_, packetID, err := GetPhantomQuestProgress(&data, client)
	if err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	if packetID != 12213 {
		t.Fatalf("expected packet id 12213, got %d", packetID)
	}

	var response protobuf.SC_12213
	tmp := client.Buffer.Bytes()
	if got := packets.GetPacketId(0, &tmp); got != 12213 {
		t.Fatalf("expected response packet 12213, got %d", got)
	}
	decodeResponse(t, client, &response)
	if len(response.GetShipCountList()) != 0 {
		t.Fatalf("expected empty ship_count_list")
	}
}

func TestGetPhantomQuestProgressDecodeFailure(t *testing.T) {
	client := setupHandlerCommander(t)
	data := []byte{0xff, 0xff, 0xff}

	client.Buffer.Reset()
	_, packetID, err := GetPhantomQuestProgress(&data, client)
	if err == nil {
		t.Fatalf("expected error")
	}
	if packetID != 12213 {
		t.Fatalf("expected packet id 12213, got %d", packetID)
	}
	if client.Buffer.Len() != 0 {
		t.Fatalf("expected no response to be written")
	}
}
