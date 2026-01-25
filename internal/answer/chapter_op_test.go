package answer

import (
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

	if err := orm.GormDB.Create(&orm.OwnedResource{
		CommanderID: client.Commander.CommanderID,
		ResourceID:  2,
		Amount:      100,
	}).Error; err != nil {
		t.Fatalf("seed oil: %v", err)
	}
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
	state, err := orm.GetChapterState(orm.GormDB, client.Commander.CommanderID)
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

	if err := orm.GormDB.Create(&orm.OwnedResource{
		CommanderID: client.Commander.CommanderID,
		ResourceID:  2,
		Amount:      100,
	}).Error; err != nil {
		t.Fatalf("seed oil: %v", err)
	}
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

	if err := orm.GormDB.Create(&orm.OwnedResource{
		CommanderID: client.Commander.CommanderID,
		ResourceID:  2,
		Amount:      100,
	}).Error; err != nil {
		t.Fatalf("seed oil: %v", err)
	}
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
	if _, err := orm.GetChapterState(orm.GormDB, client.Commander.CommanderID); err == nil {
		t.Fatalf("expected chapter state to be deleted")
	}
}

func TestChapterOpEnemyRoundUpdatesRound(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.ChapterState{})
	seedChapterTrackingConfig(t)

	if err := orm.GormDB.Create(&orm.OwnedResource{
		CommanderID: client.Commander.CommanderID,
		ResourceID:  2,
		Amount:      100,
	}).Error; err != nil {
		t.Fatalf("seed oil: %v", err)
	}
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
	state, err := orm.GetChapterState(orm.GormDB, client.Commander.CommanderID)
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
	client.Buffer.Reset()
	return err
}
