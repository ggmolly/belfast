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
	seedEliteShipTemplate(t)

	execAnswerTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(2), int64(100))
	shipA := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1001, ID: 5001}
	shipB := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1001, ID: 5002}
	execAnswerTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, create_time, change_name_timestamp) VALUES ($1, $2, $3, NOW(), NOW())", int64(shipA.ID), int64(shipA.OwnerID), int64(shipA.ShipID))
	execAnswerTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, create_time, change_name_timestamp) VALUES ($1, $2, $3, NOW(), NOW())", int64(shipB.ID), int64(shipB.OwnerID), int64(shipB.ShipID))

	if err := startChapterTracking(t, client); err != nil {
		t.Fatalf("start chapter tracking: %v", err)
	}
	state, err := orm.GetChapterState(client.Commander.CommanderID)
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
	if err := orm.UpsertChapterState(state); err != nil {
		t.Fatalf("persist chapter state: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
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

	stored, err := orm.GetChapterState(client.Commander.CommanderID)
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
	seedEliteShipTemplate(t)

	execAnswerTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(2), int64(100))
	shipA := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1001, ID: 5001}
	shipB := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1001, ID: 5002}
	execAnswerTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, create_time, change_name_timestamp) VALUES ($1, $2, $3, NOW(), NOW())", int64(shipA.ID), int64(shipA.OwnerID), int64(shipA.ShipID))
	execAnswerTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, create_time, change_name_timestamp) VALUES ($1, $2, $3, NOW(), NOW())", int64(shipB.ID), int64(shipB.OwnerID), int64(shipB.ShipID))
	if err := startChapterTracking(t, client); err != nil {
		t.Fatalf("start chapter tracking: %v", err)
	}
	state, err := orm.GetChapterState(client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load chapter state: %v", err)
	}
	fleets := []*protobuf.FLEET_INFO{{Id: proto.Uint32(1), MainTeam: []*protobuf.TEAM_INFO{{Id: proto.Uint32(1), ShipList: []uint32{5001}}}}}
	updatedState, err := setEliteFleetInState(state.State, fleets)
	if err != nil {
		t.Fatalf("set elite fleet state: %v", err)
	}
	state.State = updatedState
	if err := orm.UpsertChapterState(state); err != nil {
		t.Fatalf("persist chapter state: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
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
	seedEliteShipTemplate(t)

	execAnswerTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(2), int64(100))
	owned := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1001, ID: 5002}
	execAnswerTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, create_time, change_name_timestamp) VALUES ($1, $2, $3, NOW(), NOW())", int64(owned.ID), int64(owned.OwnerID), int64(owned.ShipID))
	if err := startChapterTracking(t, client); err != nil {
		t.Fatalf("start chapter tracking: %v", err)
	}
	state, err := orm.GetChapterState(client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load chapter state: %v", err)
	}
	fleets := []*protobuf.FLEET_INFO{{Id: proto.Uint32(1), MainTeam: []*protobuf.TEAM_INFO{{Id: proto.Uint32(1), ShipList: []uint32{5001}}}}}
	updatedState, err := setEliteFleetInState(state.State, fleets)
	if err != nil {
		t.Fatalf("set elite fleet state: %v", err)
	}
	state.State = updatedState
	if err := orm.UpsertChapterState(state); err != nil {
		t.Fatalf("persist chapter state: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
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
	seedEliteShipTemplate(t)

	ship := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1001, ID: 5001}
	execAnswerTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, create_time, change_name_timestamp) VALUES ($1, $2, $3, NOW(), NOW())", int64(ship.ID), int64(ship.OwnerID), int64(ship.ShipID))
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
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
	if _, err := orm.GetChapterState(client.Commander.CommanderID); err == nil {
		t.Fatalf("expected chapter state to remain absent")
	}
}

func TestEmptyChapterStateBlobReturnsEmptyFleetList(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.ChapterState{})
	clearTable(t, &orm.OwnedShip{})
	seedEliteShipTemplate(t)

	ship := orm.OwnedShip{OwnerID: client.Commander.CommanderID, ShipID: 1001, ID: 5001}
	execAnswerTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, create_time, change_name_timestamp) VALUES ($1, $2, $3, NOW(), NOW())", int64(ship.ID), int64(ship.OwnerID), int64(ship.ShipID))
	if err := orm.UpsertChapterState(&orm.ChapterState{CommanderID: client.Commander.CommanderID, ChapterID: 101, State: []byte{}}); err != nil {
		t.Fatalf("seed chapter state: %v", err)
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
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
	seedEliteShipTemplate(t)

	execAnswerTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(2), int64(100))
	if err := startChapterTracking(t, client); err != nil {
		t.Fatalf("start chapter tracking: %v", err)
	}
	state, err := orm.GetChapterState(client.Commander.CommanderID)
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

func seedEliteShipTemplate(t *testing.T) {
	t.Helper()
	execAnswerTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (template_id) DO NOTHING", int64(1001), "Elite Test Ship", "Elite Test Ship", int64(1), int64(1), int64(1), int64(1), int64(0))
}
