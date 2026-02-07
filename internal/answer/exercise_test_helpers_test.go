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

	commander := orm.Commander{CommanderID: 1, AccountID: 1, Level: 1, Name: "Test Commander"}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("seed commander: %v", err)
	}

	pos := uint32(999)
	for i := uint32(1); i <= 6; i++ {
		ship := orm.OwnedShip{ID: i, OwnerID: commander.CommanderID, ShipID: 100 + i, SecretaryPosition: &pos}
		if err := orm.GormDB.Create(&ship).Error; err != nil {
			t.Fatalf("seed owned ship: %v", err)
		}
	}

	fleet := orm.Fleet{
		GameID:         1,
		CommanderID:    commander.CommanderID,
		Name:           "Fleet 1",
		ShipList:       orm.Int64List{1, 2, 3, 4, 5, 6},
		MeowfficerList: orm.Int64List{},
	}
	if err := orm.GormDB.Create(&fleet).Error; err != nil {
		t.Fatalf("seed fleet 1: %v", err)
	}

	loaded := orm.Commander{CommanderID: commander.CommanderID}
	if err := loaded.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	return &connection.Client{Commander: &loaded}
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
