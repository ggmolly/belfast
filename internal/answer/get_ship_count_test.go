package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestGetShipCountEmpty(t *testing.T) {
	commander := &orm.Commander{Ships: []orm.OwnedShip{}}
	client := &connection.Client{Commander: commander}
	payload := protobuf.CS_11800{Type: proto.Uint32(0)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	_, packetID, err := GetShipCount(&buffer, client)
	if err != nil {
		t.Fatalf("get ship count failed: %v", err)
	}
	if packetID != 11801 {
		t.Fatalf("expected packet 11801, got %d", packetID)
	}

	data := client.Buffer.Bytes()
	if len(data) < 7 {
		t.Fatalf("expected buffer to include header and payload")
	}
	data = data[7:]

	var response protobuf.SC_11801
	if err := proto.Unmarshal(data, &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.GetShipCount() != 0 {
		t.Fatalf("expected ship_count 0, got %d", response.GetShipCount())
	}
}

func TestGetShipCountWithShips(t *testing.T) {
	commander := &orm.Commander{Ships: []orm.OwnedShip{{ID: 1}, {ID: 2}, {ID: 3}}}
	client := &connection.Client{Commander: commander}
	payload := protobuf.CS_11800{Type: proto.Uint32(0)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	_, packetID, err := GetShipCount(&buffer, client)
	if err != nil {
		t.Fatalf("get ship count failed: %v", err)
	}
	if packetID != 11801 {
		t.Fatalf("expected packet 11801, got %d", packetID)
	}

	data := client.Buffer.Bytes()
	if len(data) < 7 {
		t.Fatalf("expected buffer to include header and payload")
	}
	data = data[7:]

	var response protobuf.SC_11801
	if err := proto.Unmarshal(data, &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.GetShipCount() != 3 {
		t.Fatalf("expected ship_count 3, got %d", response.GetShipCount())
	}
}
