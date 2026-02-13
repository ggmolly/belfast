package answer

import (
	"os"
	"testing"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/packets"
	"google.golang.org/protobuf/proto"
)

func setupExerciseTest(t *testing.T) *connection.Client {
	t.Helper()
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearTable(t, &orm.ExerciseFleet{})
	clearTable(t, &orm.Fleet{})
	clearTable(t, &orm.OwnedShip{})
	clearTable(t, &orm.Commander{})

	if err := orm.CreateCommanderRoot(1, 1, "Test Commander", 0, 0); err != nil {
		t.Fatalf("seed commander: %v", err)
	}

	commander := orm.Commander{CommanderID: 1}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	shipIDs := make([]uint32, 0, 6)
	for i := uint32(1); i <= 6; i++ {
		ship := orm.OwnedShip{OwnerID: commander.CommanderID, ShipID: 100 + i}
		if err := ship.Create(); err != nil {
			t.Fatalf("seed owned ship: %v", err)
		}
		shipIDs = append(shipIDs, ship.ID)
	}

	if err := commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
	}
	if err := orm.CreateFleet(&commander, 1, "Fleet 1", shipIDs); err != nil {
		t.Fatalf("seed fleet 1: %v", err)
	}

	if err := commander.Load(); err != nil {
		t.Fatalf("final load commander: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func decodePacketMessage(t *testing.T, client *connection.Client, expectedPacketID int, resp proto.Message) {
	t.Helper()
	buffer := client.Buffer.Bytes()
	packetID := packets.GetPacketId(0, &buffer)
	if packetID != expectedPacketID {
		t.Fatalf("expected packet %d, got %d", expectedPacketID, packetID)
	}
	packetSize := packets.GetPacketSize(0, &buffer) + 2
	if len(buffer) < packetSize {
		t.Fatalf("expected packet size %d, got %d", packetSize, len(buffer))
	}
	payloadStart := packets.HEADER_SIZE
	payloadEnd := payloadStart + (packetSize - packets.HEADER_SIZE)
	if err := proto.Unmarshal(buffer[payloadStart:payloadEnd], resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
}
