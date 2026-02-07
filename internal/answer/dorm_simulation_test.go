package answer_test

import (
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestDormTickViaOpenAddExpPushesPopAndUpdatesShips(t *testing.T) {
	client := newDormTestClient(t)
	commanderID := client.Commander.CommanderID

	// Seed dorm template (level 1) with a 15s tick.
	seedConfigEntry(t, "ShareCfg/dorm_data_template.json", "1", `{"id":1,"capacity":40000,"consume":5,"exp":1,"time":15,"comfortable":20}`)

	shipTemplateID := uint32(time.Now().UnixNano()%1_000_000_000 + 7_000_000)
	ensureTestShipTemplate(t, shipTemplateID)

	trainID := uint32(time.Now().UnixNano()%1_000_000_000 + 30_000)
	restID := trainID + 1
	train := orm.OwnedShip{OwnerID: commanderID, ShipID: shipTemplateID, ID: trainID, State: 5, Intimacy: 5000}
	rest := orm.OwnedShip{OwnerID: commanderID, ShipID: shipTemplateID, ID: restID, State: 2, Intimacy: 5000}
	if err := orm.GormDB.Create(&train).Error; err != nil {
		t.Fatalf("failed to create training ship: %v", err)
	}
	if err := orm.GormDB.Create(&rest).Error; err != nil {
		t.Fatalf("failed to create rest ship: %v", err)
	}

	state, err := orm.GetOrCreateCommanderDormState(commanderID)
	if err != nil {
		t.Fatalf("failed to get dorm state: %v", err)
	}
	state.Level = 1
	state.Food = 100
	state.UpdatedAtUnixTimestamp = uint32(time.Now().Add(-30 * time.Second).Unix())
	if err := orm.GormDB.Save(state).Error; err != nil {
		t.Fatalf("failed to save dorm state: %v", err)
	}

	// Call OpenAddExp which should tick and push SC_19010.
	buf, err := proto.Marshal(&protobuf.CS_19015{IsOpen: proto.Uint32(1)})
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.OpenAddExp19015(&buf, client); err != nil {
		t.Fatalf("OpenAddExp19015 failed: %v", err)
	}

	// Verify pushed packet is 19010.
	buffer := client.Buffer.Bytes()
	if len(buffer) == 0 {
		t.Fatalf("expected response buffer")
	}
	packetID := packets.GetPacketId(0, &buffer)
	if packetID != 19010 {
		t.Fatalf("expected packet 19010, got %d", packetID)
	}
	packetSize := packets.GetPacketSize(0, &buffer) + 2
	payloadStart := packets.HEADER_SIZE
	payloadEnd := payloadStart + (packetSize - packets.HEADER_SIZE)
	var pop protobuf.SC_19010
	if err := proto.Unmarshal(buffer[payloadStart:payloadEnd], &pop); err != nil {
		t.Fatalf("failed to unmarshal SC_19010: %v", err)
	}
	client.Buffer.Reset()
	if len(pop.GetPopList()) != 2 {
		t.Fatalf("expected 2 pop entries, got %d", len(pop.GetPopList()))
	}

	// Two ticks (30s / 15s), 2 ships => consume 5*2*2 = 20.
	var storedState orm.CommanderDormState
	if err := orm.GormDB.Where("commander_id = ?", commanderID).First(&storedState).Error; err != nil {
		t.Fatalf("failed to reload dorm state: %v", err)
	}
	if storedState.Food != 80 {
		t.Fatalf("expected food=80, got %d", storedState.Food)
	}

	var storedTrain orm.OwnedShip
	if err := orm.GormDB.Where("owner_id = ? AND id = ?", commanderID, trainID).First(&storedTrain).Error; err != nil {
		t.Fatalf("failed to reload training ship: %v", err)
	}
	if storedTrain.StateInfo2 != 2 {
		t.Fatalf("expected training ship exp counter=2, got %d", storedTrain.StateInfo2)
	}
	if storedTrain.StateInfo3 != 2 {
		t.Fatalf("expected training ship intimacy counter=2, got %d", storedTrain.StateInfo3)
	}
	// coinPerTick = 1 + comfortable/10 = 3, ticks=2 => 6
	if storedTrain.StateInfo4 != 6 {
		t.Fatalf("expected training ship dorm icon=6, got %d", storedTrain.StateInfo4)
	}
}
