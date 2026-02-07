package answer_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func newDormTestClient(t *testing.T) *connection.Client {
	commanderID := uint32(time.Now().UnixNano())
	commander := orm.Commander{CommanderID: commanderID, AccountID: commanderID, Name: fmt.Sprintf("Dorm Commander %d", commanderID)}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func decodePacketInto(t *testing.T, client *connection.Client, expectedID int, message proto.Message) {
	buffer := client.Buffer.Bytes()
	if len(buffer) == 0 {
		t.Fatalf("expected response buffer")
	}
	packetID := packets.GetPacketId(0, &buffer)
	if packetID != expectedID {
		t.Fatalf("expected packet %d, got %d", expectedID, packetID)
	}
	packetSize := packets.GetPacketSize(0, &buffer) + 2
	payloadStart := packets.HEADER_SIZE
	payloadEnd := payloadStart + (packetSize - packets.HEADER_SIZE)
	if err := proto.Unmarshal(buffer[payloadStart:payloadEnd], message); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	client.Buffer.Reset()
}

func ensureTestShipTemplate(t *testing.T, templateID uint32) {
	ship := orm.Ship{
		TemplateID:  templateID,
		Name:        "Test Ship",
		EnglishName: "Test Ship",
		RarityID:    1,
		Star:        1,
		Type:        1,
		Nationality: 1,
		BuildTime:   1,
	}
	if err := orm.GormDB.Save(&ship).Error; err != nil {
		t.Fatalf("failed to upsert ship template: %v", err)
	}
}

func TestClaimDormIntimacyAppliesAndClears(t *testing.T) {
	client := newDormTestClient(t)
	shipTemplateID := uint32(time.Now().UnixNano()%1_000_000_000 + 5_000_000)
	ensureTestShipTemplate(t, shipTemplateID)
	ownedShipID := uint32(time.Now().UnixNano()%1_000_000_000 + 10_000)

	ship := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: shipTemplateID, ID: ownedShipID, State: 2, Intimacy: 5000, StateInfo3: 123}
	if err := orm.GormDB.Create(&ship).Error; err != nil {
		t.Fatalf("failed to create owned ship: %v", err)
	}

	payload := &protobuf.CS_19011{Id: proto.Uint32(ownedShipID)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.ClaimDormIntimacy19011(&buf, client); err != nil {
		t.Fatalf("ClaimDormIntimacy19011 failed: %v", err)
	}
	resp := &protobuf.SC_19012{}
	decodePacketInto(t, client, 19012, resp)
	if resp.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", resp.GetResult())
	}

	var stored orm.OwnedShip
	if err := orm.GormDB.Where("owner_id = ? AND id = ?", client.Commander.CommanderID, ownedShipID).First(&stored).Error; err != nil {
		t.Fatalf("failed to reload owned ship: %v", err)
	}
	if stored.Intimacy != 5123 {
		t.Fatalf("expected intimacy=5123, got %d", stored.Intimacy)
	}
	if stored.StateInfo3 != 0 {
		t.Fatalf("expected state_info_3 to be cleared")
	}
}

func TestClaimDormIntimacyAllAlsoClaimsMoney(t *testing.T) {
	client := newDormTestClient(t)
	shipTemplateID := uint32(time.Now().UnixNano()%1_000_000_000 + 6_000_000)
	ensureTestShipTemplate(t, shipTemplateID)

	ship1ID := uint32(time.Now().UnixNano()%1_000_000_000 + 20_000)
	ship2ID := ship1ID + 1
	ship1 := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: shipTemplateID, ID: ship1ID, State: 2, Intimacy: 5000, StateInfo3: 10, StateInfo4: 3}
	ship2 := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: shipTemplateID, ID: ship2ID, State: 5, Intimacy: 6000, StateInfo3: 20, StateInfo4: 7}
	if err := orm.GormDB.Create(&ship1).Error; err != nil {
		t.Fatalf("failed to create owned ship 1: %v", err)
	}
	if err := orm.GormDB.Create(&ship2).Error; err != nil {
		t.Fatalf("failed to create owned ship 2: %v", err)
	}

	var before orm.OwnedResource
	_ = orm.GormDB.Where("commander_id = ? AND resource_id = ?", client.Commander.CommanderID, 6).First(&before).Error

	payload := &protobuf.CS_19011{Id: proto.Uint32(0)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.ClaimDormIntimacy19011(&buf, client); err != nil {
		t.Fatalf("ClaimDormIntimacy19011(all) failed: %v", err)
	}
	decodePacketInto(t, client, 19012, &protobuf.SC_19012{})

	var stored1 orm.OwnedShip
	if err := orm.GormDB.Where("owner_id = ? AND id = ?", client.Commander.CommanderID, ship1ID).First(&stored1).Error; err != nil {
		t.Fatalf("failed to reload owned ship 1: %v", err)
	}
	if stored1.Intimacy != 5010 || stored1.StateInfo3 != 0 || stored1.StateInfo4 != 0 {
		t.Fatalf("unexpected ship1 after claim: intimacy=%d info3=%d info4=%d", stored1.Intimacy, stored1.StateInfo3, stored1.StateInfo4)
	}
	var stored2 orm.OwnedShip
	if err := orm.GormDB.Where("owner_id = ? AND id = ?", client.Commander.CommanderID, ship2ID).First(&stored2).Error; err != nil {
		t.Fatalf("failed to reload owned ship 2: %v", err)
	}
	if stored2.Intimacy != 6020 || stored2.StateInfo3 != 0 || stored2.StateInfo4 != 0 {
		t.Fatalf("unexpected ship2 after claim: intimacy=%d info3=%d info4=%d", stored2.Intimacy, stored2.StateInfo3, stored2.StateInfo4)
	}

	var after orm.OwnedResource
	if err := orm.GormDB.Where("commander_id = ? AND resource_id = ?", client.Commander.CommanderID, 6).First(&after).Error; err != nil {
		t.Fatalf("failed to reload resource 6: %v", err)
	}
	if after.Amount != before.Amount+10 {
		t.Fatalf("expected resource 6 to increase by 10, before=%d after=%d", before.Amount, after.Amount)
	}
}
