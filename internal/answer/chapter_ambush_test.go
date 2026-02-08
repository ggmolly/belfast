package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestChapterTrackingAmbushCellsUseAmbushFlagAndFallbackExpedition(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.ChapterState{})
	clearTable(t, &orm.ChapterProgress{})
	clearTable(t, &orm.ConfigEntry{})

	seedConfigEntry(t, "sharecfgdata/chapter_template.json", "201", `{"id":201,"grids":[[1,1,true,1],[1,2,true,5]],"ammo_total":5,"ammo_submarine":0,"group_num":1,"submarine_num":0,"support_group_num":0,"is_ambush":0,"investigation_ratio":0,"avoid_ratio":0,"chapter_strategy":[],"boss_expedition_id":[],"expedition_id_weight_list":[[101010,160,0]],"elite_expedition_list":[],"ambush_expedition_list":[],"guarder_expedition_list":[],"progress_boss":0,"oil":0,"time":100,"awards":[]}`)

	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 2, Amount: 100}).Error; err != nil {
		t.Fatalf("seed oil: %v", err)
	}

	payload := protobuf.CS_13101{
		Id: proto.Uint32(201),
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
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	var seenAmbush bool
	for _, cell := range response.GetCurrentChapter().GetCellList() {
		if cell.GetItemType() != chapterAttachAmbush {
			continue
		}
		seenAmbush = true
		if cell.GetItemFlag() != chapterCellAmbush {
			t.Fatalf("expected ambush cell flag %d, got %d", chapterCellAmbush, cell.GetItemFlag())
		}
		if cell.GetItemId() != 101010 {
			t.Fatalf("expected ambush cell to fallback to expedition 101010, got %d", cell.GetItemId())
		}
	}
	if !seenAmbush {
		t.Fatalf("expected at least one ambush cell")
	}
}

func TestChapterOpAmbushAvoidSuccessRemovesCell(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.ChapterState{})
	clearTable(t, &orm.ChapterProgress{})
	clearTable(t, &orm.OwnedShip{})
	clearTable(t, &orm.ConfigEntry{})

	// avoid_ratio=0 => dodge always succeeds.
	seedConfigEntry(t, "sharecfgdata/chapter_template.json", "202", `{"id":202,"grids":[[1,1,true,1],[1,2,true,5]],"ammo_total":5,"ammo_submarine":0,"group_num":1,"submarine_num":0,"support_group_num":0,"is_ambush":0,"investigation_ratio":0,"avoid_ratio":0,"chapter_strategy":[],"boss_expedition_id":[],"expedition_id_weight_list":[[101010,160,0]],"elite_expedition_list":[],"ambush_expedition_list":[101220],"guarder_expedition_list":[],"progress_boss":0,"oil":0,"time":100,"awards":[]}`)

	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 2, Amount: 100}).Error; err != nil {
		t.Fatalf("seed oil: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedShip{ID: 101, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 1, MaxLevel: 100, Energy: 150}).Error; err != nil {
		t.Fatalf("seed owned ship 101: %v", err)
	}

	buffer, err := proto.Marshal(&protobuf.CS_13101{Id: proto.Uint32(202), Fleet: &protobuf.FLEET_INFO{Id: proto.Uint32(1), MainTeam: []*protobuf.TEAM_INFO{{Id: proto.Uint32(1), ShipList: []uint32{101}}}}})
	if err != nil {
		t.Fatalf("marshal tracking payload: %v", err)
	}
	if _, _, err := ChapterTracking(&buffer, client); err != nil {
		t.Fatalf("chapter tracking failed: %v", err)
	}
	client.Buffer.Reset()

	movePayload := protobuf.CS_13103{Act: proto.Uint32(chapterOpMove), GroupId: proto.Uint32(1), ActArg_1: proto.Uint32(1), ActArg_2: proto.Uint32(2)}
	moveBuffer, err := proto.Marshal(&movePayload)
	if err != nil {
		t.Fatalf("marshal move payload: %v", err)
	}
	if _, _, err := ChapterOp(&moveBuffer, client); err != nil {
		t.Fatalf("chapter op move failed: %v", err)
	}
	client.Buffer.Reset()

	ambushPayload := protobuf.CS_13103{Act: proto.Uint32(chapterOpAmbush), GroupId: proto.Uint32(1), ActArg_1: proto.Uint32(1)}
	ambushBuffer, err := proto.Marshal(&ambushPayload)
	if err != nil {
		t.Fatalf("marshal ambush payload: %v", err)
	}
	if _, _, err := ChapterOp(&ambushBuffer, client); err != nil {
		t.Fatalf("chapter op ambush failed: %v", err)
	}

	var response protobuf.SC_13104
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if len(response.GetMapUpdate()) != 0 {
		t.Fatalf("expected no map updates on ambush avoid success, got %d", len(response.GetMapUpdate()))
	}

	state, err := orm.GetChapterState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	var current protobuf.CURRENTCHAPTERINFO
	if err := proto.Unmarshal(state.State, &current); err != nil {
		t.Fatalf("unmarshal state: %v", err)
	}
	if _, cell := findChapterCellAt(&current, chapterPos{Row: 1, Column: 2}); cell != nil {
		t.Fatalf("expected ambush cell to be removed from state")
	}
}

func TestChapterOpAmbushAvoidFailMarksCellActive(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.ChapterState{})
	clearTable(t, &orm.ChapterProgress{})
	clearTable(t, &orm.OwnedShip{})
	clearTable(t, &orm.ConfigEntry{})

	// avoid_ratio is large enough that chance rounds down to 0.
	seedConfigEntry(t, "sharecfgdata/chapter_template.json", "203", `{"id":203,"grids":[[1,1,true,1],[1,2,true,5]],"ammo_total":5,"ammo_submarine":0,"group_num":1,"submarine_num":0,"support_group_num":0,"is_ambush":0,"investigation_ratio":0,"avoid_ratio":1000000,"chapter_strategy":[],"boss_expedition_id":[],"expedition_id_weight_list":[[101010,160,0]],"elite_expedition_list":[],"ambush_expedition_list":[101220],"guarder_expedition_list":[],"progress_boss":0,"oil":0,"time":100,"awards":[]}`)

	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 2, Amount: 100}).Error; err != nil {
		t.Fatalf("seed oil: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedShip{ID: 101, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 1, MaxLevel: 100, Energy: 150}).Error; err != nil {
		t.Fatalf("seed owned ship 101: %v", err)
	}

	buffer, err := proto.Marshal(&protobuf.CS_13101{Id: proto.Uint32(203), Fleet: &protobuf.FLEET_INFO{Id: proto.Uint32(1), MainTeam: []*protobuf.TEAM_INFO{{Id: proto.Uint32(1), ShipList: []uint32{101}}}}})
	if err != nil {
		t.Fatalf("marshal tracking payload: %v", err)
	}
	if _, _, err := ChapterTracking(&buffer, client); err != nil {
		t.Fatalf("chapter tracking failed: %v", err)
	}
	client.Buffer.Reset()

	movePayload := protobuf.CS_13103{Act: proto.Uint32(chapterOpMove), GroupId: proto.Uint32(1), ActArg_1: proto.Uint32(1), ActArg_2: proto.Uint32(2)}
	moveBuffer, err := proto.Marshal(&movePayload)
	if err != nil {
		t.Fatalf("marshal move payload: %v", err)
	}
	if _, _, err := ChapterOp(&moveBuffer, client); err != nil {
		t.Fatalf("chapter op move failed: %v", err)
	}
	client.Buffer.Reset()

	ambushPayload := protobuf.CS_13103{Act: proto.Uint32(chapterOpAmbush), GroupId: proto.Uint32(1), ActArg_1: proto.Uint32(1)}
	ambushBuffer, err := proto.Marshal(&ambushPayload)
	if err != nil {
		t.Fatalf("marshal ambush payload: %v", err)
	}
	if _, _, err := ChapterOp(&ambushBuffer, client); err != nil {
		t.Fatalf("chapter op ambush failed: %v", err)
	}

	var response protobuf.SC_13104
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if len(response.GetMapUpdate()) != 1 {
		t.Fatalf("expected 1 map update on ambush avoid fail, got %d", len(response.GetMapUpdate()))
	}
	if response.GetMapUpdate()[0].GetItemFlag() != chapterCellActive {
		t.Fatalf("expected ambush cell flag to be active (%d), got %d", chapterCellActive, response.GetMapUpdate()[0].GetItemFlag())
	}
}

func TestChapterOpMoveTriggersAmbushWhenChanceMaxed(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.ChapterState{})
	clearTable(t, &orm.ChapterProgress{})
	clearTable(t, &orm.OwnedShip{})
	clearTable(t, &orm.ConfigEntry{})

	seedConfigEntry(t, "sharecfgdata/chapter_template.json", "301", `{"id":301,"grids":[[1,1,true,1],[1,2,true,0],[1,3,true,0],[1,4,true,0],[1,5,true,0],[1,6,true,0]],"ammo_total":5,"ammo_submarine":0,"group_num":1,"submarine_num":0,"support_group_num":0,"is_ambush":1,"investigation_ratio":999999999,"avoid_ratio":0,"chapter_strategy":[],"boss_expedition_id":[],"expedition_id_weight_list":[[101010,160,0]],"elite_expedition_list":[],"ambush_expedition_list":[101220],"guarder_expedition_list":[],"progress_boss":0,"oil":0,"time":100,"awards":[]}`)

	if err := orm.GormDB.Create(&orm.OwnedResource{CommanderID: client.Commander.CommanderID, ResourceID: 2, Amount: 100}).Error; err != nil {
		t.Fatalf("seed oil: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedShip{ID: 101, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 1, MaxLevel: 100, Energy: 150}).Error; err != nil {
		t.Fatalf("seed owned ship 101: %v", err)
	}

	buffer, err := proto.Marshal(&protobuf.CS_13101{Id: proto.Uint32(301), Fleet: &protobuf.FLEET_INFO{Id: proto.Uint32(1), MainTeam: []*protobuf.TEAM_INFO{{Id: proto.Uint32(1), ShipList: []uint32{101}}}}})
	if err != nil {
		t.Fatalf("marshal tracking payload: %v", err)
	}
	if _, _, err := ChapterTracking(&buffer, client); err != nil {
		t.Fatalf("chapter tracking failed: %v", err)
	}
	client.Buffer.Reset()

	movePayload := protobuf.CS_13103{Act: proto.Uint32(chapterOpMove), GroupId: proto.Uint32(1), ActArg_1: proto.Uint32(1), ActArg_2: proto.Uint32(6)}
	moveBuffer, err := proto.Marshal(&movePayload)
	if err != nil {
		t.Fatalf("marshal move payload: %v", err)
	}
	if _, _, err := ChapterOp(&moveBuffer, client); err != nil {
		t.Fatalf("chapter op move failed: %v", err)
	}

	var response protobuf.SC_13104
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if len(response.GetMapUpdate()) != 1 {
		t.Fatalf("expected 1 map update from ambush trigger, got %d", len(response.GetMapUpdate()))
	}
	cell := response.GetMapUpdate()[0]
	if cell.GetItemType() != chapterAttachAmbush {
		t.Fatalf("expected ambush item type %d, got %d", chapterAttachAmbush, cell.GetItemType())
	}
	if cell.GetItemFlag() != chapterCellAmbush {
		t.Fatalf("expected ambush cell flag %d, got %d", chapterCellAmbush, cell.GetItemFlag())
	}
	if cell.GetItemId() != 101220 {
		t.Fatalf("expected ambush expedition 101220, got %d", cell.GetItemId())
	}
}
