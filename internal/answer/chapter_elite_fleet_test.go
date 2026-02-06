package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestRemoveShipFromEliteFleetSuccess(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.ChapterState{})
	clearTable(t, &orm.OwnedShip{})
	seedChapterTrackingConfig(t)

	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 2, Amount: 100}).Error; err != nil {
		t.Fatalf("seed oil: %v", err)
	}
	shipA := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1, ID: 5001}
	shipB := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1, ID: 5002}
	if err := orm.GormDB.Create(&shipA).Error; err != nil {
		t.Fatalf("seed ship A: %v", err)
	}
	if err := orm.GormDB.Create(&shipB).Error; err != nil {
		t.Fatalf("seed ship B: %v", err)
	}

	if err := startChapterTracking(t, client); err != nil {
		t.Fatalf("start chapter tracking: %v", err)
	}
	state, err := orm.GetChapterState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load chapter state: %v", err)
	}

	fleets := []*protobuf.FLEET_INFO{
		{
			Id:            proto.Uint32(1),
			MainTeam:      []*protobuf.TEAM_INFO{{Id: proto.Uint32(1), ShipList: []uint32{5001, 5002}}},
			SubmarineTeam: []*protobuf.TEAM_INFO{{Id: proto.Uint32(1), ShipList: []uint32{5001}}},
			SupportTeam:   []*protobuf.TEAM_INFO{{Id: proto.Uint32(1), ShipList: []uint32{5002}}},
		},
		{
			Id:       proto.Uint32(2),
			MainTeam: []*protobuf.TEAM_INFO{{Id: proto.Uint32(1), ShipList: []uint32{5001}}},
		},
	}
	updatedState, err := setEliteFleetInState(state.State, fleets)
	if err != nil {
		t.Fatalf("set elite fleet state: %v", err)
	}
	state.State = updatedState
	if err := orm.UpsertChapterState(orm.GormDB, state); err != nil {
		t.Fatalf("persist chapter state: %v", err)
	}
	client.Buffer.Reset()

	payload := protobuf.CS_13111{ShipId: proto.Uint32(5001)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := RemoveEliteTargetShip(&buffer, client); err != nil {
		t.Fatalf("remove elite target ship: %v", err)
	}

	var response protobuf.SC_13112
	decodeResponse(t, client, &response)
	if len(response.GetFleetList()) != 2 {
		t.Fatalf("expected 2 fleets, got %d", len(response.GetFleetList()))
	}
	if got := response.GetFleetList()[0].GetMainTeam()[0].GetShipList(); len(got) != 1 || got[0] != 5002 {
		t.Fatalf("expected ship 5001 removed from main team, got %v", got)
	}
	if got := response.GetFleetList()[0].GetSubmarineTeam()[0].GetShipList(); len(got) != 0 {
		t.Fatalf("expected ship 5001 removed from submarine team, got %v", got)
	}
	if got := response.GetFleetList()[1].GetMainTeam()[0].GetShipList(); len(got) != 0 {
		t.Fatalf("expected ship 5001 removed from second fleet, got %v", got)
	}

	stored, err := orm.GetChapterState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("reload chapter state: %v", err)
	}
	storedFleets, err := parseEliteFleetFromState(stored.State)
	if err != nil {
		t.Fatalf("parse stored fleets: %v", err)
	}
	if got := storedFleets[0].GetMainTeam()[0].GetShipList(); len(got) != 1 || got[0] != 5002 {
		t.Fatalf("expected stored fleet updated, got %v", got)
	}
}

func TestRemoveShipNotInEliteFleet(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.ChapterState{})
	clearTable(t, &orm.OwnedShip{})
	seedChapterTrackingConfig(t)

	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 2, Amount: 100}).Error; err != nil {
		t.Fatalf("seed oil: %v", err)
	}
	shipA := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1, ID: 5001}
	shipB := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1, ID: 5002}
	if err := orm.GormDB.Create(&shipA).Error; err != nil {
		t.Fatalf("seed ship: %v", err)
	}
	if err := orm.GormDB.Create(&shipB).Error; err != nil {
		t.Fatalf("seed ship: %v", err)
	}
	if err := startChapterTracking(t, client); err != nil {
		t.Fatalf("start chapter tracking: %v", err)
	}
	state, err := orm.GetChapterState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load chapter state: %v", err)
	}
	fleets := []*protobuf.FLEET_INFO{{Id: proto.Uint32(1), MainTeam: []*protobuf.TEAM_INFO{{Id: proto.Uint32(1), ShipList: []uint32{5001}}}}}
	updatedState, err := setEliteFleetInState(state.State, fleets)
	if err != nil {
		t.Fatalf("set elite fleet state: %v", err)
	}
	state.State = updatedState
	if err := orm.UpsertChapterState(orm.GormDB, state); err != nil {
		t.Fatalf("persist chapter state: %v", err)
	}
	client.Buffer.Reset()

	payload := protobuf.CS_13111{ShipId: proto.Uint32(5002)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := RemoveEliteTargetShip(&buffer, client); err != nil {
		t.Fatalf("remove elite target ship: %v", err)
	}

	var response protobuf.SC_13112
	decodeResponse(t, client, &response)
	if len(response.GetFleetList()) != 1 {
		t.Fatalf("expected 1 fleet, got %d", len(response.GetFleetList()))
	}
	if got := response.GetFleetList()[0].GetMainTeam()[0].GetShipList(); len(got) != 1 || got[0] != 5001 {
		t.Fatalf("expected fleet list unchanged, got %v", got)
	}
}

func TestRemoveShipNotOwned(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.ChapterState{})
	clearTable(t, &orm.OwnedShip{})
	seedChapterTrackingConfig(t)

	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 2, Amount: 100}).Error; err != nil {
		t.Fatalf("seed oil: %v", err)
	}
	owned := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1, ID: 5002}
	if err := orm.GormDB.Create(&owned).Error; err != nil {
		t.Fatalf("seed ship: %v", err)
	}
	if err := startChapterTracking(t, client); err != nil {
		t.Fatalf("start chapter tracking: %v", err)
	}
	state, err := orm.GetChapterState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load chapter state: %v", err)
	}
	fleets := []*protobuf.FLEET_INFO{{Id: proto.Uint32(1), MainTeam: []*protobuf.TEAM_INFO{{Id: proto.Uint32(1), ShipList: []uint32{5001}}}}}
	updatedState, err := setEliteFleetInState(state.State, fleets)
	if err != nil {
		t.Fatalf("set elite fleet state: %v", err)
	}
	state.State = updatedState
	if err := orm.UpsertChapterState(orm.GormDB, state); err != nil {
		t.Fatalf("persist chapter state: %v", err)
	}
	client.Buffer.Reset()

	payload := protobuf.CS_13111{ShipId: proto.Uint32(5001)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := RemoveEliteTargetShip(&buffer, client); err == nil {
		t.Fatalf("expected error")
	}
	if client.Buffer.Len() != 0 {
		t.Fatalf("expected no response to be written")
	}
}

func TestNoEliteFleetState(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.ChapterState{})
	clearTable(t, &orm.OwnedShip{})

	ship := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1, ID: 5001}
	if err := orm.GormDB.Create(&ship).Error; err != nil {
		t.Fatalf("seed ship: %v", err)
	}
	client.Buffer.Reset()

	payload := protobuf.CS_13111{ShipId: proto.Uint32(5001)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := RemoveEliteTargetShip(&buffer, client); err != nil {
		t.Fatalf("remove elite target ship: %v", err)
	}

	var response protobuf.SC_13112
	decodeResponse(t, client, &response)
	if len(response.GetFleetList()) != 0 {
		t.Fatalf("expected empty fleet list, got %d", len(response.GetFleetList()))
	}
	if _, err := orm.GetChapterState(orm.GormDB, client.Commander.CommanderID); err == nil {
		t.Fatalf("expected chapter state to remain absent")
	}
}

func TestEmptyChapterStateBlobReturnsEmptyFleetList(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.ChapterState{})
	clearTable(t, &orm.OwnedShip{})

	ship := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1, ID: 5001}
	if err := orm.GormDB.Create(&ship).Error; err != nil {
		t.Fatalf("seed ship: %v", err)
	}
	if err := orm.GormDB.Create(&orm.ChapterState{CommanderID: client.Commander.CommanderID, ChapterID: 101, State: []byte{}}).Error; err != nil {
		t.Fatalf("seed chapter state: %v", err)
	}
	client.Buffer.Reset()

	payload := protobuf.CS_13111{ShipId: proto.Uint32(5001)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := RemoveEliteTargetShip(&buffer, client); err != nil {
		t.Fatalf("remove elite target ship: %v", err)
	}

	var response protobuf.SC_13112
	decodeResponse(t, client, &response)
	if len(response.GetFleetList()) != 0 {
		t.Fatalf("expected empty fleet list, got %d", len(response.GetFleetList()))
	}
}

func TestParseEliteFleetFromState(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.ChapterState{})
	seedChapterTrackingConfig(t)

	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 2, Amount: 100}).Error; err != nil {
		t.Fatalf("seed oil: %v", err)
	}
	if err := startChapterTracking(t, client); err != nil {
		t.Fatalf("start chapter tracking: %v", err)
	}
	state, err := orm.GetChapterState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load chapter state: %v", err)
	}

	input := []*protobuf.FLEET_INFO{{Id: proto.Uint32(7), MainTeam: []*protobuf.TEAM_INFO{{Id: proto.Uint32(1), ShipList: []uint32{1, 2, 3}}}}}
	withElite, err := setEliteFleetInState(state.State, input)
	if err != nil {
		t.Fatalf("set elite fleet state: %v", err)
	}

	parsed, err := parseEliteFleetFromState(withElite)
	if err != nil {
		t.Fatalf("parse elite fleet: %v", err)
	}
	if len(parsed) != 1 {
		t.Fatalf("expected 1 fleet, got %d", len(parsed))
	}
	if parsed[0].GetId() != 7 {
		t.Fatalf("expected fleet id 7, got %d", parsed[0].GetId())
	}
	if got := parsed[0].GetMainTeam()[0].GetShipList(); len(got) != 3 || got[0] != 1 || got[2] != 3 {
		t.Fatalf("unexpected ship list: %v", got)
	}
}
