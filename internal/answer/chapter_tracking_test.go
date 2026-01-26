package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestChapterTrackingSuccess(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.ChapterState{})
	clearTable(t, &orm.ChapterProgress{})
	seedChapterTrackingConfig(t)

	if err := orm.GormDB.Create(&orm.OwnedResource{
		CommanderID: client.Commander.CommanderID,
		ResourceID:  2,
		Amount:      100,
	}).Error; err != nil {
		t.Fatalf("seed oil: %v", err)
	}
	if err := orm.GormDB.Create(&orm.CommanderItem{
		CommanderID: client.Commander.CommanderID,
		ItemID:      20001,
		Count:       1,
	}).Error; err != nil {
		t.Fatalf("seed item: %v", err)
	}

	payload := protobuf.CS_13101{
		Id: proto.Uint32(101),
		Fleet: &protobuf.FLEET_INFO{
			Id: proto.Uint32(1),
			MainTeam: []*protobuf.TEAM_INFO{
				{Id: proto.Uint32(1), ShipList: []uint32{101}},
			},
		},
		OperationItem: proto.Uint32(20001),
	}
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
	if response.GetCurrentChapter().GetId() != 101 {
		t.Fatalf("expected chapter id 101, got %d", response.GetCurrentChapter().GetId())
	}
	if len(response.GetCurrentChapter().GetMainGroupList()) != 1 {
		t.Fatalf("expected 1 main group, got %d", len(response.GetCurrentChapter().GetMainGroupList()))
	}
	if len(response.GetCurrentChapter().GetOperationBuff()) != 1 || response.GetCurrentChapter().GetOperationBuff()[0] != 2 {
		t.Fatalf("expected operation buff 2")
	}
	for _, cell := range response.GetCurrentChapter().GetCellList() {
		if cell.GetItemType() == 6 && cell.GetItemId() == 0 {
			t.Fatalf("expected enemy cell to include item_id")
		}
	}

	var state orm.ChapterState
	if err := orm.GormDB.First(&state, "commander_id = ?", client.Commander.CommanderID).Error; err != nil {
		t.Fatalf("chapter state missing: %v", err)
	}
	if _, err := orm.GetChapterProgress(orm.GormDB, client.Commander.CommanderID, 101); err != nil {
		t.Fatalf("chapter progress missing: %v", err)
	}
	var oil orm.OwnedResource
	if err := orm.GormDB.First(&oil, "commander_id = ? AND resource_id = ?", client.Commander.CommanderID, 2).Error; err != nil {
		t.Fatalf("load oil: %v", err)
	}
	if oil.Amount != 88 {
		t.Fatalf("expected oil 88, got %d", oil.Amount)
	}
	var item orm.CommanderItem
	if err := orm.GormDB.First(&item, "commander_id = ? AND item_id = ?", client.Commander.CommanderID, 20001).Error; err != nil {
		t.Fatalf("load item: %v", err)
	}
	if item.Count != 0 {
		t.Fatalf("expected item count 0, got %d", item.Count)
	}
}

func TestChapterTrackingInvalidChapter(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.ChapterState{})
	clearTable(t, &orm.ChapterProgress{})

	if err := orm.GormDB.Create(&orm.OwnedResource{
		CommanderID: client.Commander.CommanderID,
		ResourceID:  2,
		Amount:      100,
	}).Error; err != nil {
		t.Fatalf("seed oil: %v", err)
	}
	payload := protobuf.CS_13101{
		Id: proto.Uint32(999),
		Fleet: &protobuf.FLEET_INFO{
			Id: proto.Uint32(1),
			MainTeam: []*protobuf.TEAM_INFO{
				{Id: proto.Uint32(1), ShipList: []uint32{101}},
			},
		},
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ChapterTracking(&buffer, client); err != nil {
		t.Fatalf("chapter tracking failed: %v", err)
	}
	var response protobuf.SC_13102
	decodeResponse(t, client, &response)
	if response.GetResult() != 1 {
		t.Fatalf("expected result 1, got %d", response.GetResult())
	}
	var state orm.ChapterState
	if err := orm.GormDB.First(&state, "commander_id = ?", client.Commander.CommanderID).Error; err == nil {
		t.Fatalf("expected no chapter state")
	}
}

func seedChapterTrackingConfig(t *testing.T) {
	seedConfigEntry(t, "sharecfgdata/chapter_template.json", "101", `{"id":101,"grids":[[1,1,true,1],[1,2,true,6],[1,3,true,8]],"ammo_total":5,"ammo_submarine":2,"group_num":1,"submarine_num":0,"support_group_num":0,"chapter_strategy":[1016],"boss_expedition_id":[9001],"expedition_id_weight_list":[[101010,160,0]],"elite_expedition_list":[101210],"ambush_expedition_list":[101220],"guarder_expedition_list":[101100],"star_require_1":1,"num_1":1,"star_require_2":2,"num_2":1,"star_require_3":4,"num_3":3,"progress_boss":100,"oil":10,"time":100}`)
	seedConfigEntry(t, "sharecfgdata/item_data_statistics.json", "20001", `{"id":20001,"usage_arg":[1]}`)
	seedConfigEntry(t, "ShareCfg/benefit_buff_template.json", "1", `{"id":1,"benefit_type":"more_oil","benefit_effect":"20","benefit_condition":"0"}`)
	seedConfigEntry(t, "ShareCfg/benefit_buff_template.json", "2", `{"id":2,"benefit_type":"desc","benefit_effect":"0","benefit_condition":"20001"}`)
}
