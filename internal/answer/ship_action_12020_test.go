package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestShipAction12020Success(t *testing.T) {
	commander := &orm.Commander{
		OwnedShipsMap: map[uint32]*orm.OwnedShip{
			10: {ID: 10},
		},
	}
	client := &connection.Client{Commander: commander}
	payload := protobuf.CS_12020{ShipId: proto.Uint32(10)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	_, packetID, err := ShipAction12020(&buffer, client)
	if err != nil {
		t.Fatalf("ship action failed: %v", err)
	}
	if packetID != 12021 {
		t.Fatalf("expected packet 12021, got %d", packetID)
	}

	data := client.Buffer.Bytes()
	if len(data) < 7 {
		t.Fatalf("expected buffer to include header and payload")
	}
	data = data[7:]

	var response protobuf.SC_12021
	if err := proto.Unmarshal(data, &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
}

func TestShipAction12020MissingShip(t *testing.T) {
	commander := &orm.Commander{OwnedShipsMap: map[uint32]*orm.OwnedShip{}}
	client := &connection.Client{Commander: commander}
	payload := protobuf.CS_12020{ShipId: proto.Uint32(99)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	_, packetID, err := ShipAction12020(&buffer, client)
	if err != nil {
		t.Fatalf("ship action failed: %v", err)
	}
	if packetID != 12021 {
		t.Fatalf("expected packet 12021, got %d", packetID)
	}

	data := client.Buffer.Bytes()
	if len(data) < 7 {
		t.Fatalf("expected buffer to include header and payload")
	}
	data = data[7:]

	var response protobuf.SC_12021
	if err := proto.Unmarshal(data, &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.GetResult() != 1 {
		t.Fatalf("expected result 1, got %d", response.GetResult())
	}
}

func TestShipAction12020BadPayload(t *testing.T) {
	commander := &orm.Commander{OwnedShipsMap: map[uint32]*orm.OwnedShip{}}
	client := &connection.Client{Commander: commander}
	buffer := []byte{0xff}

	_, packetID, err := ShipAction12020(&buffer, client)
	if err == nil {
		t.Fatalf("expected unmarshal error")
	}
	if packetID != 12021 {
		t.Fatalf("expected packet 12021, got %d", packetID)
	}
}
