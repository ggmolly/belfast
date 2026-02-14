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
	if err := orm.CreateCommanderRoot(commanderID, commanderID, fmt.Sprintf("Dorm Commander %d", commanderID), 0, 0); err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	commander := orm.Commander{CommanderID: commanderID}
	if err := commander.Load(); err != nil {
		t.Fatalf("failed to load commander: %v", err)
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
	execAnswerExternalTestSQLT(t, `
INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (template_id)
DO UPDATE SET
	name = EXCLUDED.name,
	english_name = EXCLUDED.english_name,
	rarity_id = EXCLUDED.rarity_id,
	star = EXCLUDED.star,
	type = EXCLUDED.type,
	nationality = EXCLUDED.nationality,
	build_time = EXCLUDED.build_time
`, templateID, "Test Ship", "Test Ship", 1, 1, 1, 1, 1)
}

func TestClaimDormIntimacyAppliesAndClears(t *testing.T) {
	client := newDormTestClient(t)
	shipTemplateID := uint32(time.Now().UnixNano()%1_000_000_000 + 5_000_000)
	ensureTestShipTemplate(t, shipTemplateID)
	ownedShipID := uint32(time.Now().UnixNano()%1_000_000_000 + 10_000)

	execAnswerExternalTestSQLT(t, `
INSERT INTO owned_ships (owner_id, ship_id, id, state, intimacy, state_info3)
VALUES ($1, $2, $3, $4, $5, $6)
`, client.Commander.CommanderID, shipTemplateID, ownedShipID, 2, 5000, 123)

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

	stored, err := orm.GetOwnedShipByOwnerAndID(client.Commander.CommanderID, ownedShipID)
	if err != nil {
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
	execAnswerExternalTestSQLT(t, `
INSERT INTO owned_ships (owner_id, ship_id, id, state, intimacy, state_info3, state_info4)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`, client.Commander.CommanderID, shipTemplateID, ship1ID, 2, 5000, 10, 3)
	execAnswerExternalTestSQLT(t, `
INSERT INTO owned_ships (owner_id, ship_id, id, state, intimacy, state_info3, state_info4)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`, client.Commander.CommanderID, shipTemplateID, ship2ID, 5, 6000, 20, 7)

	before := queryAnswerExternalTestInt64(t, `
SELECT COALESCE((
	SELECT amount
	FROM owned_resources
	WHERE commander_id = $1 AND resource_id = $2
), 0)
`, client.Commander.CommanderID, 6)

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

	stored1, err := orm.GetOwnedShipByOwnerAndID(client.Commander.CommanderID, ship1ID)
	if err != nil {
		t.Fatalf("failed to reload owned ship 1: %v", err)
	}
	if stored1.Intimacy != 5010 || stored1.StateInfo3 != 0 || stored1.StateInfo4 != 0 {
		t.Fatalf("unexpected ship1 after claim: intimacy=%d info3=%d info4=%d", stored1.Intimacy, stored1.StateInfo3, stored1.StateInfo4)
	}
	stored2, err := orm.GetOwnedShipByOwnerAndID(client.Commander.CommanderID, ship2ID)
	if err != nil {
		t.Fatalf("failed to reload owned ship 2: %v", err)
	}
	if stored2.Intimacy != 6020 || stored2.StateInfo3 != 0 || stored2.StateInfo4 != 0 {
		t.Fatalf("unexpected ship2 after claim: intimacy=%d info3=%d info4=%d", stored2.Intimacy, stored2.StateInfo3, stored2.StateInfo4)
	}

	after := queryAnswerExternalTestInt64(t, `
SELECT COALESCE((
	SELECT amount
	FROM owned_resources
	WHERE commander_id = $1 AND resource_id = $2
), 0)
`, client.Commander.CommanderID, 6)
	if after != before+10 {
		t.Fatalf("expected resource 6 to increase by 10, before=%d after=%d", before, after)
	}
}
