package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestChapterTrackingAllowsNegativeAmbushRatioExtraValues(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.ChapterState{})
	clearTable(t, &orm.ChapterProgress{})
	clearTable(t, &orm.ConfigEntry{})

	seedConfigEntry(t, "sharecfgdata/chapter_template.json", "901", `{"id":901,"grids":[[1,1,true,1],[1,2,true,5]],"ammo_total":5,"ammo_submarine":0,"group_num":1,"submarine_num":0,"support_group_num":0,"is_ambush":0,"investigation_ratio":0,"avoid_ratio":0,"ambush_ratio_extra":[[-20000]],"chapter_strategy":[],"boss_expedition_id":[],"expedition_id_weight_list":[[101010,160,0]],"elite_expedition_list":[],"ambush_expedition_list":[],"guarder_expedition_list":[],"progress_boss":0,"oil":0,"time":100,"awards":[]}`)

	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 2, Amount: 100}).Error; err != nil {
		t.Fatalf("seed oil: %v", err)
	}

	payload := protobuf.CS_13101{Id: proto.Uint32(901), Fleet: &protobuf.FLEET_INFO{Id: proto.Uint32(1), MainTeam: []*protobuf.TEAM_INFO{{Id: proto.Uint32(1), ShipList: []uint32{101}}}}}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ChapterTracking(&buffer, client); err != nil {
		t.Fatalf("chapter tracking failed: %v", err)
	}

	var response protobuf.SC_13102
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
}
