package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
)

func TestCommanderFleetIncludesName(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	client.Commander.Fleets = []orm.Fleet{
		{
			GameID:   1,
			Name:     "Alpha",
			ShipList: orm.Int64List{101, 102},
		},
	}

	buffer := []byte{}
	if _, _, err := CommanderFleet(&buffer, client); err != nil {
		t.Fatalf("commander fleet failed: %v", err)
	}

	var response protobuf.SC_12101
	decodeResponse(t, client, &response)
	if len(response.GetGroupList()) != 1 {
		t.Fatalf("expected 1 group, got %d", len(response.GetGroupList()))
	}
	if response.GetGroupList()[0].GetName() != "Alpha" {
		t.Fatalf("expected group name Alpha, got %q", response.GetGroupList()[0].GetName())
	}
}
