package answer

import (
	"fmt"
	"testing"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestChapterOpMoveUpdatesState(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.ChapterState{})
	seedChapterTrackingConfig(t)

	execAnswerTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(2), int64(100))
	if err := startChapterTracking(t, client); err != nil {
		t.Fatalf("start tracking: %v", err)
	}
	payload := protobuf.CS_13103{
		Act:      proto.Uint32(1),
		GroupId:  proto.Uint32(1),
		ActArg_1: proto.Uint32(1),
		ActArg_2: proto.Uint32(2),
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ChapterOp(&buffer, client); err != nil {
		t.Fatalf("chapter op failed: %v", err)
	}
	var response protobuf.SC_13104
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if len(response.GetMovePath()) != 2 {
		t.Fatalf("expected move path length 2, got %d", len(response.GetMovePath()))
	}
	state, err := orm.GetChapterState(client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	var current protobuf.CURRENTCHAPTERINFO
	if err := proto.Unmarshal(state.State, &current); err != nil {
		t.Fatalf("unmarshal state: %v", err)
	}
	group := current.GetMainGroupList()[0]
	if group.GetPos().GetColumn() != 2 {
		t.Fatalf("expected group column 2, got %d", group.GetPos().GetColumn())
	}
}

func TestChapterOpRequestReturnsState(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.ChapterState{})
	seedChapterTrackingConfig(t)

	execAnswerTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(2), int64(100))
	if err := startChapterTracking(t, client); err != nil {
		t.Fatalf("start tracking: %v", err)
	}
	payload := protobuf.CS_13103{
		Act:     proto.Uint32(49),
		GroupId: proto.Uint32(1),
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ChapterOp(&buffer, client); err != nil {
		t.Fatalf("chapter op failed: %v", err)
	}
	var response protobuf.SC_13104
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if len(response.GetMapUpdate()) != 3 {
		t.Fatalf("expected 3 map updates, got %d", len(response.GetMapUpdate()))
	}
}

func TestChapterOpRetreatClearsState(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.ChapterState{})
	seedChapterTrackingConfig(t)

	execAnswerTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(2), int64(100))
	if err := startChapterTracking(t, client); err != nil {
		t.Fatalf("start tracking: %v", err)
	}
	payload := protobuf.CS_13103{
		Act:     proto.Uint32(0),
		GroupId: proto.Uint32(1),
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ChapterOp(&buffer, client); err != nil {
		t.Fatalf("chapter op failed: %v", err)
	}
	if _, err := orm.GetChapterState(client.Commander.CommanderID); err == nil {
		t.Fatalf("expected chapter state to be deleted")
	}
}

func TestChapterOpEnemyRoundUpdatesRound(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.ChapterState{})
	seedChapterTrackingConfig(t)

	execAnswerTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(2), int64(100))
	if err := startChapterTracking(t, client); err != nil {
		t.Fatalf("start tracking: %v", err)
	}
	payload := protobuf.CS_13103{
		Act:     proto.Uint32(8),
		GroupId: proto.Uint32(1),
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ChapterOp(&buffer, client); err != nil {
		t.Fatalf("chapter op failed: %v", err)
	}
	var response protobuf.SC_13104
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	state, err := orm.GetChapterState(client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	var current protobuf.CURRENTCHAPTERINFO
	if err := proto.Unmarshal(state.State, &current); err != nil {
		t.Fatalf("unmarshal state: %v", err)
	}
	if current.GetRound() != 1 {
		t.Fatalf("expected round 1, got %d", current.GetRound())
	}
}

func startChapterTracking(t *testing.T, client *connection.Client) error {
	if err := prepareChapterTrackingClient(t, client); err != nil {
		return err
	}
	if queryAnswerTestInt64(t, "SELECT COUNT(*) FROM config_entries WHERE category = $1 AND key = $2", "sharecfgdata/chapter_template.json", "101") == 0 {
		return fmt.Errorf("missing chapter config entry")
	}
	if !client.Commander.HasEnoughResource(2, 1) {
		return fmt.Errorf("commander oil still empty: %d", client.Commander.GetResourceCount(2))
	}

	payload := protobuf.CS_13101{
		Id: proto.Uint32(101),
		Fleet: &protobuf.FLEET_INFO{
			Id: proto.Uint32(1),
			MainTeam: []*protobuf.TEAM_INFO{
				{Id: proto.Uint32(1), ShipList: []uint32{101}},
			},
		},
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		return err
	}
	_, _, err = ChapterTracking(&buffer, client)
	if err != nil {
		return err
	}
	var response protobuf.SC_13102
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		client.Buffer.Reset()
		return fmt.Errorf("chapter tracking result %d", response.GetResult())
	}
	client.Buffer.Reset()
	return nil
}

func prepareChapterTrackingClient(t *testing.T, client *connection.Client) error {
	t.Helper()
	ensureChapterTrackingShip(t, client)
	if err := client.Commander.AddResource(2, 100); err != nil {
		return err
	}
	if client.Commander.CommanderItemsMap == nil {
		client.Commander.CommanderItemsMap = make(map[uint32]*orm.CommanderItem)
	}
	if client.Commander.MiscItemsMap == nil {
		client.Commander.MiscItemsMap = make(map[uint32]*orm.CommanderMiscItem)
	}
	return nil
}

func ensureChapterTrackingShip(t *testing.T, client *connection.Client) {
	t.Helper()
	if client.Commander.OwnedShipsMap == nil {
		client.Commander.OwnedShipsMap = make(map[uint32]*orm.OwnedShip)
	}
	if _, ok := client.Commander.OwnedShipsMap[101]; ok {
		return
	}
	execAnswerTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, level, max_level, energy, create_time, change_name_timestamp) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW()) ON CONFLICT (id) DO NOTHING", int64(101), int64(client.Commander.CommanderID), int64(1001), int64(1), int64(100), int64(150))
	client.Commander.OwnedShipsMap[101] = &orm.OwnedShip{ID: 101, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 1, MaxLevel: 100, Energy: 150}
}
