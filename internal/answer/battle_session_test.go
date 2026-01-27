package answer

import (
	"errors"
	"reflect"
	"testing"

	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func TestBeginStageCreatesBattleSession(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.BattleSession{})

	payload := protobuf.CS_40001{
		System:     proto.Uint32(1),
		ShipIdList: []uint32{101, 102},
		Data:       proto.Uint32(3001),
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := BeginStage(&buffer, client); err != nil {
		t.Fatalf("begin stage failed: %v", err)
	}
	var response protobuf.SC_40002
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if response.GetKey() == 0 {
		t.Fatalf("expected non-zero key")
	}
	session, err := orm.GetBattleSession(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("get battle session: %v", err)
	}
	if session.System != 1 || session.StageID != 3001 {
		t.Fatalf("unexpected session values")
	}
	if session.Key != response.GetKey() {
		t.Fatalf("expected session key %d, got %d", response.GetKey(), session.Key)
	}
	if !reflect.DeepEqual(session.ShipIDs, orm.Int64List{101, 102}) {
		t.Fatalf("unexpected ship ids: %v", session.ShipIDs)
	}
}

func TestFinishStageClearsBattleSession(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.BattleSession{})
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
	if err := startChapterTracking(t, client); err != nil {
		t.Fatalf("start tracking: %v", err)
	}

	beginPayload := protobuf.CS_40001{
		System:     proto.Uint32(1),
		ShipIdList: []uint32{101, 102},
		// use an expedition id that exists in seeded chapter config
		Data: proto.Uint32(101010),
	}
	beginBuffer, err := proto.Marshal(&beginPayload)
	if err != nil {
		t.Fatalf("marshal begin payload: %v", err)
	}
	if _, _, err := BeginStage(&beginBuffer, client); err != nil {
		t.Fatalf("begin stage failed: %v", err)
	}
	var beginResponse protobuf.SC_40002
	decodeResponse(t, client, &beginResponse)
	client.Buffer.Reset()

	finishPayload := protobuf.CS_40003{
		System:         proto.Uint32(1),
		Data:           proto.Uint32(101010),
		Key:            proto.Uint32(beginResponse.GetKey()),
		TotalTime:      proto.Uint32(1),
		BotPercentage:  proto.Uint32(0),
		ExtraParam:     proto.Uint32(0),
		AutoBefore:     proto.Uint32(0),
		AutoSwitchTime: proto.Uint32(0),
		AutoAfter:      proto.Uint32(0),
	}
	finishBuffer, err := proto.Marshal(&finishPayload)
	if err != nil {
		t.Fatalf("marshal finish payload: %v", err)
	}
	if _, _, err := FinishStage(&finishBuffer, client); err != nil {
		t.Fatalf("finish stage failed: %v", err)
	}
	var finishResponse protobuf.SC_40004
	decodeResponse(t, client, &finishResponse)
	if finishResponse.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", finishResponse.GetResult())
	}
	if finishResponse.GetMvp() != 101 {
		t.Fatalf("expected mvp 101, got %d", finishResponse.GetMvp())
	}
	if len(finishResponse.GetShipExpList()) != 2 {
		t.Fatalf("expected 2 ship exp entries, got %d", len(finishResponse.GetShipExpList()))
	}
	_, err = orm.GetBattleSession(orm.GormDB, client.Commander.CommanderID)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected session to be deleted, got %v", err)
	}
	state, err := orm.GetChapterState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		t.Fatalf("load chapter state: %v", err)
	}
	var current protobuf.CURRENTCHAPTERINFO
	if err := proto.Unmarshal(state.State, &current); err != nil {
		t.Fatalf("unmarshal state: %v", err)
	}
	found := false
	for _, cell := range current.GetCellList() {
		if cell.GetItemId() == 101010 {
			if cell.GetItemFlag() != chapterCellDisabled {
				t.Fatalf("expected cell flag disabled for defeated enemy")
			}
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected enemy cell to be present in state")
	}
	progress, err := orm.GetChapterProgress(orm.GormDB, client.Commander.CommanderID, 101)
	if err != nil {
		t.Fatalf("load chapter progress: %v", err)
	}
	if progress.KillEnemyCount != 1 {
		t.Fatalf("expected kill enemy count 1, got %d", progress.KillEnemyCount)
	}
}

func TestQuitBattleClearsBattleSession(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.BattleSession{})

	beginPayload := protobuf.CS_40001{
		System: proto.Uint32(2),
		Data:   proto.Uint32(4001),
	}
	beginBuffer, err := proto.Marshal(&beginPayload)
	if err != nil {
		t.Fatalf("marshal begin payload: %v", err)
	}
	if _, _, err := BeginStage(&beginBuffer, client); err != nil {
		t.Fatalf("begin stage failed: %v", err)
	}
	client.Buffer.Reset()

	payload := protobuf.CS_40005{System: proto.Uint32(2)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := QuitBattle(&buffer, client); err != nil {
		t.Fatalf("quit battle failed: %v", err)
	}
	var response protobuf.SC_40006
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	_, err = orm.GetBattleSession(orm.GormDB, client.Commander.CommanderID)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected session to be deleted, got %v", err)
	}
}

func TestFinishStageUpdatesBossProgress(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.BattleSession{})
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
	if err := startChapterTracking(t, client); err != nil {
		t.Fatalf("start tracking: %v", err)
	}

	beginPayload := protobuf.CS_40001{
		System:     proto.Uint32(1),
		ShipIdList: []uint32{101},
		// use boss expedition id from seeded chapter config
		Data: proto.Uint32(9001),
	}
	beginBuffer, err := proto.Marshal(&beginPayload)
	if err != nil {
		t.Fatalf("marshal begin payload: %v", err)
	}
	if _, _, err := BeginStage(&beginBuffer, client); err != nil {
		t.Fatalf("begin stage failed: %v", err)
	}
	var beginResponse protobuf.SC_40002
	decodeResponse(t, client, &beginResponse)
	client.Buffer.Reset()

	finishPayload := protobuf.CS_40003{
		System:         proto.Uint32(1),
		Data:           proto.Uint32(9001),
		Key:            proto.Uint32(beginResponse.GetKey()),
		Score:          proto.Uint32(4),
		TotalTime:      proto.Uint32(1),
		BotPercentage:  proto.Uint32(0),
		ExtraParam:     proto.Uint32(0),
		AutoBefore:     proto.Uint32(0),
		AutoSwitchTime: proto.Uint32(0),
		AutoAfter:      proto.Uint32(0),
	}
	finishBuffer, err := proto.Marshal(&finishPayload)
	if err != nil {
		t.Fatalf("marshal finish payload: %v", err)
	}
	if _, _, err := FinishStage(&finishBuffer, client); err != nil {
		t.Fatalf("finish stage failed: %v", err)
	}
	progress, err := orm.GetChapterProgress(orm.GormDB, client.Commander.CommanderID, 101)
	if err != nil {
		t.Fatalf("load chapter progress: %v", err)
	}
	if progress.Progress != 100 {
		t.Fatalf("expected progress 100, got %d", progress.Progress)
	}
	if progress.PassCount != 1 {
		t.Fatalf("expected pass count 1, got %d", progress.PassCount)
	}
	if progress.KillBossCount != 1 {
		t.Fatalf("expected kill boss count 1, got %d", progress.KillBossCount)
	}
}

func TestFinishStageGrantsChapterAwards(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.BattleSession{})
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.ChapterState{})
	clearTable(t, &orm.ChapterProgress{})
	clearTable(t, &orm.CommanderItem{})

	seedConfigEntry(t, "sharecfgdata/chapter_template.json", "202", `{"id":202,"grids":[[1,1,true,1],[1,2,true,8]],"ammo_total":5,"ammo_submarine":2,"group_num":1,"submarine_num":0,"support_group_num":0,"chapter_strategy":[],"boss_expedition_id":[9001],"expedition_id_weight_list":[[101010,160,0]],"elite_expedition_list":[101210],"ambush_expedition_list":[101220],"guarder_expedition_list":[101100],"progress_boss":100,"oil":10,"time":100,"awards":[[2,8000]]}`)
	item := orm.Item{ID: 8000, Name: "Test Item", Rarity: 1, ShopID: -2, Type: 1, VirtualType: 0}
	if err := orm.GormDB.FirstOrCreate(&item, orm.Item{ID: 8000}).Error; err != nil {
		t.Fatalf("seed item: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedResource{
		CommanderID: client.Commander.CommanderID,
		ResourceID:  2,
		Amount:      100,
	}).Error; err != nil {
		t.Fatalf("seed oil: %v", err)
	}

	trackingPayload := protobuf.CS_13101{
		Id: proto.Uint32(202),
		Fleet: &protobuf.FLEET_INFO{
			Id: proto.Uint32(1),
			MainTeam: []*protobuf.TEAM_INFO{
				{Id: proto.Uint32(1), ShipList: []uint32{101}},
			},
		},
	}
	trackingBuffer, err := proto.Marshal(&trackingPayload)
	if err != nil {
		t.Fatalf("marshal tracking payload: %v", err)
	}
	if _, _, err := ChapterTracking(&trackingBuffer, client); err != nil {
		t.Fatalf("chapter tracking failed: %v", err)
	}
	client.Buffer.Reset()

	beginPayload := protobuf.CS_40001{
		System:     proto.Uint32(1),
		ShipIdList: []uint32{101},
		Data:       proto.Uint32(9001),
	}
	beginBuffer, err := proto.Marshal(&beginPayload)
	if err != nil {
		t.Fatalf("marshal begin payload: %v", err)
	}
	if _, _, err := BeginStage(&beginBuffer, client); err != nil {
		t.Fatalf("begin stage failed: %v", err)
	}
	var beginResponse protobuf.SC_40002
	decodeResponse(t, client, &beginResponse)
	client.Buffer.Reset()

	finishPayload := protobuf.CS_40003{
		System:         proto.Uint32(1),
		Data:           proto.Uint32(9001),
		Key:            proto.Uint32(beginResponse.GetKey()),
		Score:          proto.Uint32(4),
		TotalTime:      proto.Uint32(1),
		BotPercentage:  proto.Uint32(0),
		ExtraParam:     proto.Uint32(0),
		AutoBefore:     proto.Uint32(0),
		AutoSwitchTime: proto.Uint32(0),
		AutoAfter:      proto.Uint32(0),
	}
	finishBuffer, err := proto.Marshal(&finishPayload)
	if err != nil {
		t.Fatalf("marshal finish payload: %v", err)
	}
	if _, _, err := FinishStage(&finishBuffer, client); err != nil {
		t.Fatalf("finish stage failed: %v", err)
	}
	var finishResponse protobuf.SC_40004
	decodeResponse(t, client, &finishResponse)
	if len(finishResponse.GetDropInfo()) != 1 {
		t.Fatalf("expected 1 drop, got %d", len(finishResponse.GetDropInfo()))
	}
	drop := finishResponse.GetDropInfo()[0]
	if drop.GetType() != 2 || drop.GetId() != 8000 || drop.GetNumber() != 1 {
		t.Fatalf("unexpected drop: %+v", drop)
	}
	var owned orm.CommanderItem
	if err := orm.GormDB.First(&owned, "commander_id = ? AND item_id = ?", client.Commander.CommanderID, 8000).Error; err != nil {
		t.Fatalf("load awarded item: %v", err)
	}
	if owned.Count != 1 {
		t.Fatalf("expected item count 1, got %d", owned.Count)
	}
}

func TestFinishStageResolvesVirtualAwardDrops(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.BattleSession{})
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.ChapterState{})
	clearTable(t, &orm.ChapterProgress{})
	clearTable(t, &orm.CommanderItem{})
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.Ship{})

	ship := orm.Ship{
		TemplateID:  101061,
		Name:        "Test Ship",
		RarityID:    1,
		Star:        1,
		Type:        1,
		Nationality: 1,
		BuildTime:   0,
	}
	if err := orm.GormDB.Create(&ship).Error; err != nil {
		t.Fatalf("seed ship: %v", err)
	}

	seedConfigEntry(t, "sharecfgdata/item_virtual_data_statistics.json", "90001", `{"id":90001,"type":99,"virtual_type":0,"display_icon":[[4,101061,1]]}`)
	seedConfigEntry(t, "sharecfgdata/chapter_template.json", "203", `{"id":203,"grids":[[1,1,true,1],[1,2,true,8]],"ammo_total":5,"ammo_submarine":2,"group_num":1,"submarine_num":0,"support_group_num":0,"chapter_strategy":[],"boss_expedition_id":[9002],"expedition_id_weight_list":[[101010,160,0]],"elite_expedition_list":[101210],"ambush_expedition_list":[101220],"guarder_expedition_list":[101100],"progress_boss":100,"oil":10,"time":100,"awards":[[2,90001]]}`)
	if err := orm.GormDB.Create(&orm.OwnedResource{
		CommanderID: client.Commander.CommanderID,
		ResourceID:  2,
		Amount:      100,
	}).Error; err != nil {
		t.Fatalf("seed oil: %v", err)
	}

	trackingPayload := protobuf.CS_13101{
		Id: proto.Uint32(203),
		Fleet: &protobuf.FLEET_INFO{
			Id: proto.Uint32(1),
			MainTeam: []*protobuf.TEAM_INFO{
				{Id: proto.Uint32(1), ShipList: []uint32{101}},
			},
		},
	}
	trackingBuffer, err := proto.Marshal(&trackingPayload)
	if err != nil {
		t.Fatalf("marshal tracking payload: %v", err)
	}
	if _, _, err := ChapterTracking(&trackingBuffer, client); err != nil {
		t.Fatalf("chapter tracking failed: %v", err)
	}
	client.Buffer.Reset()

	beginPayload := protobuf.CS_40001{
		System:     proto.Uint32(1),
		ShipIdList: []uint32{101},
		Data:       proto.Uint32(9002),
	}
	beginBuffer, err := proto.Marshal(&beginPayload)
	if err != nil {
		t.Fatalf("marshal begin payload: %v", err)
	}
	if _, _, err := BeginStage(&beginBuffer, client); err != nil {
		t.Fatalf("begin stage failed: %v", err)
	}
	var beginResponse protobuf.SC_40002
	decodeResponse(t, client, &beginResponse)
	client.Buffer.Reset()

	finishPayload := protobuf.CS_40003{
		System:         proto.Uint32(1),
		Data:           proto.Uint32(9002),
		Key:            proto.Uint32(beginResponse.GetKey()),
		Score:          proto.Uint32(4),
		TotalTime:      proto.Uint32(1),
		BotPercentage:  proto.Uint32(0),
		ExtraParam:     proto.Uint32(0),
		AutoBefore:     proto.Uint32(0),
		AutoSwitchTime: proto.Uint32(0),
		AutoAfter:      proto.Uint32(0),
	}
	finishBuffer, err := proto.Marshal(&finishPayload)
	if err != nil {
		t.Fatalf("marshal finish payload: %v", err)
	}
	if _, _, err := FinishStage(&finishBuffer, client); err != nil {
		t.Fatalf("finish stage failed: %v", err)
	}
	var finishResponse protobuf.SC_40004
	decodeResponse(t, client, &finishResponse)
	if len(finishResponse.GetDropInfo()) != 1 {
		t.Fatalf("expected 1 drop, got %d", len(finishResponse.GetDropInfo()))
	}
	drop := finishResponse.GetDropInfo()[0]
	if drop.GetType() != consts.DROP_TYPE_SHIP || drop.GetId() != 101061 || drop.GetNumber() != 1 {
		t.Fatalf("unexpected drop: %+v", drop)
	}
	var owned orm.OwnedShip
	if err := orm.GormDB.First(&owned, "owner_id = ? AND ship_id = ?", client.Commander.CommanderID, 101061).Error; err != nil {
		t.Fatalf("load awarded ship: %v", err)
	}
}

func TestThirdClearKeepsRawStarCounts(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.BattleSession{})
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
	if err := startChapterTracking(t, client); err != nil {
		t.Fatalf("start tracking: %v", err)
	}
	progress := &orm.ChapterProgress{
		CommanderID:    client.Commander.CommanderID,
		ChapterID:      101,
		Progress:       100,
		KillEnemyCount: 1,
		PassCount:      2,
	}
	if err := orm.UpsertChapterProgress(orm.GormDB, progress); err != nil {
		t.Fatalf("seed chapter progress: %v", err)
	}

	beginPayload := protobuf.CS_40001{
		System:     proto.Uint32(1),
		ShipIdList: []uint32{101},
		Data:       proto.Uint32(9001),
	}
	beginBuffer, err := proto.Marshal(&beginPayload)
	if err != nil {
		t.Fatalf("marshal begin payload: %v", err)
	}
	if _, _, err := BeginStage(&beginBuffer, client); err != nil {
		t.Fatalf("begin stage failed: %v", err)
	}
	var beginResponse protobuf.SC_40002
	decodeResponse(t, client, &beginResponse)
	client.Buffer.Reset()

	finishPayload := protobuf.CS_40003{
		System:         proto.Uint32(1),
		Data:           proto.Uint32(9001),
		Key:            proto.Uint32(beginResponse.GetKey()),
		Score:          proto.Uint32(4),
		TotalTime:      proto.Uint32(1),
		BotPercentage:  proto.Uint32(0),
		ExtraParam:     proto.Uint32(0),
		AutoBefore:     proto.Uint32(0),
		AutoSwitchTime: proto.Uint32(0),
		AutoAfter:      proto.Uint32(0),
	}
	finishBuffer, err := proto.Marshal(&finishPayload)
	if err != nil {
		t.Fatalf("marshal finish payload: %v", err)
	}
	if _, _, err := FinishStage(&finishBuffer, client); err != nil {
		t.Fatalf("finish stage failed: %v", err)
	}

	updated, err := orm.GetChapterProgress(orm.GormDB, client.Commander.CommanderID, 101)
	if err != nil {
		t.Fatalf("load chapter progress: %v", err)
	}
	if updated.PassCount != 3 {
		t.Fatalf("expected pass count 3, got %d", updated.PassCount)
	}
	if updated.KillBossCount != 1 {
		t.Fatalf("expected kill boss count 1, got %d", updated.KillBossCount)
	}
	if updated.KillEnemyCount != 1 {
		t.Fatalf("expected kill enemy count 1, got %d", updated.KillEnemyCount)
	}
}

func TestDailyQuickBattleReturnsRewards(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	payload := protobuf.CS_40007{
		System: proto.Uint32(1),
		Id:     proto.Uint32(9001),
		Cnt:    proto.Uint32(2),
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := DailyQuickBattle(&buffer, client); err != nil {
		t.Fatalf("daily quick battle failed: %v", err)
	}
	var response protobuf.SC_40008
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if len(response.GetRewardList()) != 2 {
		t.Fatalf("expected 2 rewards, got %d", len(response.GetRewardList()))
	}
}

func TestFinishStageAppliesExpMoraleAndCommanderExp(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.BattleSession{})
	clearTable(t, &orm.OwnedShip{})
	clearTable(t, &orm.Ship{})
	clearTable(t, &orm.ConfigEntry{})

	seedConfigEntry(t, "sharecfgdata/expedition_data_template.json", "101010", `{"id":101010,"exp":100,"level":10}`)
	seedConfigEntry(t, "ShareCfg/ship_level.json", "1", `{"level":1,"exp":100,"exp_ur":120}`)
	seedConfigEntry(t, "ShareCfg/user_level.json", "30", `{"level":30,"exp":100}`)

	if err := orm.GormDB.Create(&orm.Ship{TemplateID: 1001, Name: "Test DD", EnglishName: "Test DD", RarityID: 3, Star: 1, Type: 1, Nationality: 1}).Error; err != nil {
		t.Fatalf("seed ship 1001: %v", err)
	}
	if err := orm.GormDB.Create(&orm.Ship{TemplateID: 1002, Name: "Test CL", EnglishName: "Test CL", RarityID: 3, Star: 1, Type: 2, Nationality: 1}).Error; err != nil {
		t.Fatalf("seed ship 1002: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedShip{ID: 101, OwnerID: client.Commander.CommanderID, ShipID: 1001, Level: 1, MaxLevel: 100, Energy: 150}).Error; err != nil {
		t.Fatalf("seed owned ship 101: %v", err)
	}
	if err := orm.GormDB.Create(&orm.OwnedShip{ID: 102, OwnerID: client.Commander.CommanderID, ShipID: 1002, Level: 1, MaxLevel: 100, Energy: 150}).Error; err != nil {
		t.Fatalf("seed owned ship 102: %v", err)
	}

	beginPayload := protobuf.CS_40001{
		System:     proto.Uint32(1),
		ShipIdList: []uint32{101, 102},
		Data:       proto.Uint32(101010),
	}
	beginBuffer, err := proto.Marshal(&beginPayload)
	if err != nil {
		t.Fatalf("marshal begin payload: %v", err)
	}
	if _, _, err := BeginStage(&beginBuffer, client); err != nil {
		t.Fatalf("begin stage failed: %v", err)
	}
	var beginResponse protobuf.SC_40002
	decodeResponse(t, client, &beginResponse)
	client.Buffer.Reset()

	finishPayload := protobuf.CS_40003{
		System:    proto.Uint32(1),
		Data:      proto.Uint32(101010),
		Key:       proto.Uint32(beginResponse.GetKey()),
		Score:     proto.Uint32(4),
		TotalTime: proto.Uint32(1),
		Statistics: []*protobuf.STATISTICSINFO{
			{ShipId: proto.Uint32(101), DamageCause: proto.Uint32(0), DamageCaused: proto.Uint32(100), HpRest: proto.Uint32(100), MaxDamageOnce: proto.Uint32(100), ShipGearScore: proto.Uint32(0)},
			{ShipId: proto.Uint32(102), DamageCause: proto.Uint32(0), DamageCaused: proto.Uint32(200), HpRest: proto.Uint32(100), MaxDamageOnce: proto.Uint32(200), ShipGearScore: proto.Uint32(0)},
		},
		BotPercentage:  proto.Uint32(0),
		ExtraParam:     proto.Uint32(0),
		AutoBefore:     proto.Uint32(0),
		AutoSwitchTime: proto.Uint32(0),
		AutoAfter:      proto.Uint32(0),
	}
	finishBuffer, err := proto.Marshal(&finishPayload)
	if err != nil {
		t.Fatalf("marshal finish payload: %v", err)
	}
	if _, _, err := FinishStage(&finishBuffer, client); err != nil {
		t.Fatalf("finish stage failed: %v", err)
	}
	var finishResponse protobuf.SC_40004
	decodeResponse(t, client, &finishResponse)
	if finishResponse.GetPlayerExp() != 24 {
		t.Fatalf("expected player exp 24, got %d", finishResponse.GetPlayerExp())
	}
	if finishResponse.GetMvp() != 102 {
		t.Fatalf("expected mvp 102, got %d", finishResponse.GetMvp())
	}
	if len(finishResponse.GetShipExpList()) != 2 {
		t.Fatalf("expected 2 ship exp entries, got %d", len(finishResponse.GetShipExpList()))
	}
	shipExpMap := map[uint32]*protobuf.SHIP_EXP{}
	for _, entry := range finishResponse.GetShipExpList() {
		shipExpMap[entry.GetShipId()] = entry
	}
	if shipExpMap[101].GetExp() != 216 {
		t.Fatalf("expected ship 101 exp 216, got %d", shipExpMap[101].GetExp())
	}
	if shipExpMap[102].GetExp() != 288 {
		t.Fatalf("expected ship 102 exp 288, got %d", shipExpMap[102].GetExp())
	}
	if shipExpMap[101].GetEnergy() != 2 || shipExpMap[101].GetIntimacy() != 10100 {
		t.Fatalf("expected ship 101 energy 2 intimacy 10100, got %d %d", shipExpMap[101].GetEnergy(), shipExpMap[101].GetIntimacy())
	}

	var owned orm.OwnedShip
	if err := orm.GormDB.First(&owned, "id = ?", 101).Error; err != nil {
		t.Fatalf("load ship 101: %v", err)
	}
	if owned.Energy != 148 {
		t.Fatalf("expected ship 101 energy 148, got %d", owned.Energy)
	}
	if owned.Intimacy != 5100 {
		t.Fatalf("expected ship 101 intimacy 5100, got %d", owned.Intimacy)
	}
	if owned.Level != 2 || owned.Exp != 116 {
		t.Fatalf("expected ship 101 level 2 exp 116, got %d %d", owned.Level, owned.Exp)
	}
	owned = orm.OwnedShip{}
	if err := orm.GormDB.First(&owned, "id = ?", 102).Error; err != nil {
		t.Fatalf("load ship 102: %v", err)
	}
	if owned.Energy != 148 {
		t.Fatalf("expected ship 102 energy 148, got %d", owned.Energy)
	}
	if owned.Intimacy != 5100 {
		t.Fatalf("expected ship 102 intimacy 5100, got %d", owned.Intimacy)
	}
}
