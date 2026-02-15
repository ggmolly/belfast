package answer_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func newVisitBackyardRequesterWithTarget(t *testing.T) (*connection.Client, uint32, string) {
	t.Helper()
	orm.InitDatabase()

	requesterID := uint32(time.Now().UnixNano())
	targetID := requesterID + 1
	targetName := fmt.Sprintf("Visit Target %d", targetID)

	if err := orm.CreateCommanderRoot(requesterID, requesterID, fmt.Sprintf("Requester %d", requesterID), 0, 0); err != nil {
		t.Fatalf("failed to create requester: %v", err)
	}
	if err := orm.CreateCommanderRoot(targetID, targetID, targetName, 0, 0); err != nil {
		t.Fatalf("failed to create target: %v", err)
	}

	requester := orm.Commander{CommanderID: requesterID}
	if err := requester.Load(); err != nil {
		t.Fatalf("failed to load requester: %v", err)
	}

	return &connection.Client{Commander: &requester}, targetID, targetName
}

func TestVisitBackyard19101Success(t *testing.T) {
	client, targetID, targetName := newVisitBackyardRequesterWithTarget(t)
	shipTemplateID := uint32(time.Now().UnixNano()%1_000_000_000 + 7_000_000)
	ensureTestShipTemplate(t, shipTemplateID)
	targetShipID := uint32(time.Now().UnixNano()%1_000_000_000 + 30_000)

	execAnswerExternalTestSQLT(t, `
INSERT INTO commander_dorm_states (
	commander_id,
	level,
	food,
	food_max_increase_count,
	floor_num,
	exp_pos,
	next_timestamp,
	load_exp,
	load_food,
	load_time,
	updated_at_unix_timestamp
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (commander_id)
DO UPDATE SET
	level = EXCLUDED.level,
	food = EXCLUDED.food,
	food_max_increase_count = EXCLUDED.food_max_increase_count,
	floor_num = EXCLUDED.floor_num,
	exp_pos = EXCLUDED.exp_pos,
	next_timestamp = EXCLUDED.next_timestamp,
	load_exp = EXCLUDED.load_exp,
	load_food = EXCLUDED.load_food,
	load_time = EXCLUDED.load_time,
	updated_at_unix_timestamp = EXCLUDED.updated_at_unix_timestamp
`, targetID, 2, 33, 4, 2, 3, 0, 0, 0, 0, 0)

	execAnswerExternalTestSQLT(t, `
INSERT INTO commander_furnitures (commander_id, furniture_id, count, get_time)
VALUES ($1, $2, $3, $4)
ON CONFLICT (commander_id, furniture_id)
DO UPDATE SET count = EXCLUDED.count, get_time = EXCLUDED.get_time
`, targetID, 900001, 2, 1700000000)

	execAnswerExternalTestSQLT(t, `
INSERT INTO commander_dorm_floor_layouts (commander_id, floor, furniture_put_list)
VALUES ($1, $2, $3::jsonb)
ON CONFLICT (commander_id, floor)
DO UPDATE SET furniture_put_list = EXCLUDED.furniture_put_list
`, targetID, 1, `[{"id":"f-1","x":2,"y":4,"dir":1,"child":[],"parent":0,"shipId":0}]`)

	execAnswerExternalTestSQLT(t, `
INSERT INTO owned_ships (owner_id, ship_id, id, state, skin_id)
VALUES ($1, $2, $3, $4, $5)
`, targetID, shipTemplateID, targetShipID, 2, 1234)

	request := &protobuf.CS_19101{UserId: proto.Uint32(targetID)}
	buffer, err := proto.Marshal(request)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	client.Buffer.Reset()
	if _, _, err := answer.VisitBackyard19101(&buffer, client); err != nil {
		t.Fatalf("VisitBackyard19101 failed: %v", err)
	}

	response := &protobuf.SC_19102{}
	decodePacketInto(t, client, 19102, response)

	if response.GetLv() == 0 {
		t.Fatalf("expected non-zero dorm level")
	}
	if response.GetName() != targetName {
		t.Fatalf("expected target name %q, got %q", targetName, response.GetName())
	}
	if response.GetFloorNum() != 2 {
		t.Fatalf("expected floor_num=2, got %d", response.GetFloorNum())
	}
	if len(response.GetShipIdList()) != 1 {
		t.Fatalf("expected 1 dorm ship, got %d", len(response.GetShipIdList()))
	}
	ship := response.GetShipIdList()[0]
	if ship.GetId() != targetShipID || ship.GetTid() != shipTemplateID || ship.GetState() != 2 || ship.GetSkinId() != 1234 {
		t.Fatalf("unexpected ship projection id=%d tid=%d state=%d skin=%d", ship.GetId(), ship.GetTid(), ship.GetState(), ship.GetSkinId())
	}
	if len(response.GetFurnitureIdList()) != 1 {
		t.Fatalf("expected 1 furniture info, got %d", len(response.GetFurnitureIdList()))
	}
	if len(response.GetFurniturePutList()) != 1 {
		t.Fatalf("expected 1 floor layout, got %d", len(response.GetFurniturePutList()))
	}
}

func TestVisitBackyard19101EmptyDorm(t *testing.T) {
	client, targetID, targetName := newVisitBackyardRequesterWithTarget(t)

	request := &protobuf.CS_19101{UserId: proto.Uint32(targetID)}
	buffer, err := proto.Marshal(request)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	client.Buffer.Reset()
	if _, _, err := answer.VisitBackyard19101(&buffer, client); err != nil {
		t.Fatalf("VisitBackyard19101 failed: %v", err)
	}

	response := &protobuf.SC_19102{}
	decodePacketInto(t, client, 19102, response)

	if response.GetLv() == 0 {
		t.Fatalf("expected non-zero level for existing target")
	}
	if response.GetName() != targetName {
		t.Fatalf("expected target name %q, got %q", targetName, response.GetName())
	}
	if len(response.GetShipIdList()) != 0 {
		t.Fatalf("expected empty ship list, got %d", len(response.GetShipIdList()))
	}
	if len(response.GetFurnitureIdList()) != 0 {
		t.Fatalf("expected empty furniture list, got %d", len(response.GetFurnitureIdList()))
	}
	if len(response.GetFurniturePutList()) != 0 {
		t.Fatalf("expected empty floor layouts, got %d", len(response.GetFurniturePutList()))
	}
}

func TestVisitBackyard19101MissingTargetReturnsUnavailable(t *testing.T) {
	client := newDormTestClient(t)

	request := &protobuf.CS_19101{UserId: proto.Uint32(999999999)}
	buffer, err := proto.Marshal(request)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	client.Buffer.Reset()
	if _, _, err := answer.VisitBackyard19101(&buffer, client); err != nil {
		t.Fatalf("expected fallback response, got error: %v", err)
	}

	response := &protobuf.SC_19102{}
	decodePacketInto(t, client, 19102, response)
	if response.GetLv() != 0 {
		t.Fatalf("expected lv=0 fallback for missing target, got %d", response.GetLv())
	}
}

func TestVisitBackyard19101MalformedPayload(t *testing.T) {
	client := newDormTestClient(t)
	malformed := []byte{0x80}

	client.Buffer.Reset()
	if _, _, err := answer.VisitBackyard19101(&malformed, client); err == nil {
		t.Fatalf("expected unmarshal error")
	}
	if client.Buffer.Len() != 0 {
		t.Fatalf("expected no response buffer on malformed payload")
	}
}
