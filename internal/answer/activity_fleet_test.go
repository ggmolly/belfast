package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestEditActivityFleetSuccess(t *testing.T) {
	client := setupConfigTest(t)
	clearTable(t, &orm.ActivityFleet{})
	seedConfigEntry(t, "ShareCfg/activity_template.json", "10", `{"id":10}`)

	client.Commander.OwnedShipsMap = map[uint32]*orm.OwnedShip{
		1: {ID: 1, OwnerID: client.Commander.CommanderID, ShipID: 100},
	}

	request := protobuf.CS_11204{
		ActivityId: proto.Uint32(10),
		GroupList: []*protobuf.GROUPINFO_P11{
			{
				Id:       proto.Uint32(1),
				ShipList: []uint32{1},
				Commanders: []*protobuf.COMMANDERSINFO{
					{Pos: proto.Uint32(1), Id: proto.Uint32(99)},
				},
			},
		},
	}
	data, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request failed: %v", err)
	}

	buffer := data
	if _, _, err := EditActivityFleet(&buffer, client); err != nil {
		t.Fatalf("edit activity fleet failed: %v", err)
	}

	var response protobuf.SC_11205
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if response.GetActivityId() != 10 {
		t.Fatalf("expected activity id 10, got %d", response.GetActivityId())
	}

	groups, found, err := orm.LoadActivityFleetGroups(client.Commander.CommanderID, 10)
	if err != nil {
		t.Fatalf("load activity fleet groups failed: %v", err)
	}
	if !found {
		t.Fatalf("expected activity fleet groups to be stored")
	}
	if len(groups) != 1 || groups[0].ID != 1 {
		t.Fatalf("unexpected groups: %v", groups)
	}
}

func TestEditActivityFleetInvalidShip(t *testing.T) {
	client := setupConfigTest(t)
	clearTable(t, &orm.ActivityFleet{})
	seedConfigEntry(t, "ShareCfg/activity_template.json", "10", `{"id":10}`)

	client.Commander.OwnedShipsMap = map[uint32]*orm.OwnedShip{}

	request := protobuf.CS_11204{
		ActivityId: proto.Uint32(10),
		GroupList: []*protobuf.GROUPINFO_P11{
			{
				Id:       proto.Uint32(1),
				ShipList: []uint32{999},
			},
		},
	}
	data, err := proto.Marshal(&request)
	if err != nil {
		t.Fatalf("marshal request failed: %v", err)
	}

	buffer := data
	if _, _, err := EditActivityFleet(&buffer, client); err != nil {
		t.Fatalf("edit activity fleet failed: %v", err)
	}

	var response protobuf.SC_11205
	decodeResponse(t, client, &response)
	if response.GetResult() != 1 {
		t.Fatalf("expected result 1, got %d", response.GetResult())
	}

	_, found, err := orm.LoadActivityFleetGroups(client.Commander.CommanderID, 10)
	if err != nil {
		t.Fatalf("load activity fleet groups failed: %v", err)
	}
	if found {
		t.Fatalf("expected activity fleet groups to be absent")
	}
}

func TestActivitiesIncludeStoredActivityFleetGroups(t *testing.T) {
	client := setupConfigTest(t)
	clearTable(t, &orm.ActivityFleet{})
	seedConfigEntry(t, "ShareCfg/activity_template.json", "1", `{"id":1,"time":["timer",[[2024,1,1],[0,0,0]],[[2024,1,2],[0,0,0]]]}`)
	seedActivityAllowlist(t, []uint32{1})

	groups := orm.ActivityFleetGroupList{
		{
			ID:       5,
			ShipList: []uint32{1, 2},
			Commanders: []orm.ActivityFleetCommander{
				{Pos: 1, ID: 77},
			},
		},
	}
	if err := orm.SaveActivityFleetGroups(client.Commander.CommanderID, 1, groups); err != nil {
		t.Fatalf("save activity fleet groups failed: %v", err)
	}

	buffer := []byte{}
	if _, _, err := Activities(&buffer, client); err != nil {
		t.Fatalf("activities failed: %v", err)
	}

	var response protobuf.SC_11200
	decodeResponse(t, client, &response)
	if len(response.GetActivityList()) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(response.GetActivityList()))
	}
	activity := response.GetActivityList()[0]
	if len(activity.GetGroupList()) != 1 || activity.GetGroupList()[0].GetId() != 5 {
		t.Fatalf("expected stored group list to be included")
	}
}
