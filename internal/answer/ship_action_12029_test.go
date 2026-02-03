package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestShipAction12029Success(t *testing.T) {
	commander := &orm.Commander{
		OwnedShipsMap: map[uint32]*orm.OwnedShip{
			10: {ID: 10},
		},
	}
	client := &connection.Client{Commander: commander}
	payload := protobuf.CS_12029{Id: proto.Uint32(1), Num: proto.Uint32(2)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	_, packetID, err := ShipAction12029(&buffer, client)
	if err != nil {
		t.Fatalf("ship action failed: %v", err)
	}
	if packetID != 12030 {
		t.Fatalf("expected packet 12030, got %d", packetID)
	}
	if len(commander.OwnedShipsMap) != 1 {
		t.Fatalf("expected owned ships unchanged, got %d", len(commander.OwnedShipsMap))
	}

	data := client.Buffer.Bytes()
	if len(data) < 7 {
		t.Fatalf("expected buffer to include header and payload")
	}
	data = data[7:]

	var response protobuf.SC_12030
	if err := proto.Unmarshal(data, &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if len(response.GetShipList()) != 0 {
		t.Fatalf("expected empty ship list, got %d", len(response.GetShipList()))
	}
}

func TestShipAction12029BadPayload(t *testing.T) {
	commander := &orm.Commander{OwnedShipsMap: map[uint32]*orm.OwnedShip{}}
	client := &connection.Client{Commander: commander}
	buffer := []byte{0xff}

	_, packetID, err := ShipAction12029(&buffer, client)
	if err == nil {
		t.Fatalf("expected unmarshal error")
	}
	if packetID != 12030 {
		t.Fatalf("expected packet 12030, got %d", packetID)
	}
}
