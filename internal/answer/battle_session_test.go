package answer

import (
	"errors"
	"reflect"
	"testing"

	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
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
	session, err := orm.GetBattleSession(client.Commander.CommanderID)
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

	execAnswerTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(2), int64(100))
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
	_, err = orm.GetBattleSession(client.Commander.CommanderID)
	if !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("expected session to be deleted, got %v", err)
	}
	state, err := orm.GetChapterState(client.Commander.CommanderID)
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
	progress, err := orm.GetChapterProgress(client.Commander.CommanderID, 101)
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
	_, err = orm.GetBattleSession(client.Commander.CommanderID)
	if !errors.Is(err, db.ErrNotFound) {
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

	execAnswerTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(2), int64(100))
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
	progress, err := orm.GetChapterProgress(client.Commander.CommanderID, 101)
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
	execAnswerTestSQLT(t, "INSERT INTO items (id, name, rarity, shop_id, type, virtual_type) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (id) DO NOTHING", int64(8000), "Test Item", int64(1), int64(-2), int64(1), int64(0))
	execAnswerTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(2), int64(100))
	ensureChapterTrackingShip(t, client)
	if err := client.Commander.AddResource(2, 100); err != nil {
		t.Fatalf("add oil: %v", err)
	}
	if client.Commander.CommanderItemsMap == nil {
		client.Commander.CommanderItemsMap = make(map[uint32]*orm.CommanderItem)
	}
	if client.Commander.MiscItemsMap == nil {
		client.Commander.MiscItemsMap = make(map[uint32]*orm.CommanderMiscItem)
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
	var trackingResponse protobuf.SC_13102
	decodeResponse(t, client, &trackingResponse)
	if trackingResponse.GetResult() != 0 {
		t.Fatalf("expected chapter tracking result 0, got %d", trackingResponse.GetResult())
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
	owned := queryAnswerTestInt64(t, "SELECT count FROM commander_items WHERE commander_id = $1 AND item_id = $2", int64(client.Commander.CommanderID), int64(8000))
	if owned != 1 {
		t.Fatalf("expected item count 1, got %d", owned)
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
	execAnswerTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(ship.TemplateID), ship.Name, ship.Name, int64(ship.RarityID), int64(ship.Star), int64(ship.Type), int64(ship.Nationality), int64(ship.BuildTime))

	seedConfigEntry(t, "sharecfgdata/item_virtual_data_statistics.json", "90001", `{"id":90001,"type":99,"virtual_type":0,"display_icon":[[4,101061,1]]}`)
	seedConfigEntry(t, "sharecfgdata/chapter_template.json", "203", `{"id":203,"grids":[[1,1,true,1],[1,2,true,8]],"ammo_total":5,"ammo_submarine":2,"group_num":1,"submarine_num":0,"support_group_num":0,"chapter_strategy":[],"boss_expedition_id":[9002],"expedition_id_weight_list":[[101010,160,0]],"elite_expedition_list":[101210],"ambush_expedition_list":[101220],"guarder_expedition_list":[101100],"progress_boss":100,"oil":10,"time":100,"awards":[[2,90001]]}`)
	execAnswerTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(2), int64(100))
	ensureChapterTrackingShip(t, client)
	if err := client.Commander.AddResource(2, 100); err != nil {
		t.Fatalf("add oil: %v", err)
	}
	if client.Commander.CommanderItemsMap == nil {
		client.Commander.CommanderItemsMap = make(map[uint32]*orm.CommanderItem)
	}
	if client.Commander.MiscItemsMap == nil {
		client.Commander.MiscItemsMap = make(map[uint32]*orm.CommanderMiscItem)
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
	var trackingResponse protobuf.SC_13102
	decodeResponse(t, client, &trackingResponse)
	if trackingResponse.GetResult() != 0 {
		t.Fatalf("expected chapter tracking result 0, got %d", trackingResponse.GetResult())
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
	ownedCount := queryAnswerTestInt64(t, "SELECT COUNT(*) FROM owned_ships WHERE owner_id = $1 AND ship_id = $2", int64(client.Commander.CommanderID), int64(101061))
	if ownedCount == 0 {
		t.Fatalf("load awarded ship: expected row")
	}
}

func TestThirdClearKeepsRawStarCounts(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.BattleSession{})
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.ChapterState{})
	clearTable(t, &orm.ChapterProgress{})
	seedChapterTrackingConfig(t)

	execAnswerTestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(client.Commander.CommanderID), int64(2), int64(100))
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
	if err := orm.UpsertChapterProgress(progress); err != nil {
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

	updated, err := orm.GetChapterProgress(client.Commander.CommanderID, 101)
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

	execAnswerTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(1001), "Test DD", "Test DD", int64(3), int64(1), int64(1), int64(1), int64(0))
	execAnswerTestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(1002), "Test CL", "Test CL", int64(3), int64(1), int64(2), int64(1), int64(0))
	execAnswerTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, level, max_level, energy, create_time, change_name_timestamp) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())", int64(101), int64(client.Commander.CommanderID), int64(1001), int64(1), int64(100), int64(150))
	execAnswerTestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, level, max_level, energy, create_time, change_name_timestamp) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())", int64(102), int64(client.Commander.CommanderID), int64(1002), int64(1), int64(100), int64(150))
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("reload commander: %v", err)
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

	owned, err := orm.GetOwnedShipByOwnerAndID(client.Commander.CommanderID, 101)
	if err != nil {
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
	owned, err = orm.GetOwnedShipByOwnerAndID(client.Commander.CommanderID, 102)
	if err != nil {
		t.Fatalf("load ship 102: %v", err)
	}
	if owned.Energy != 148 {
		t.Fatalf("expected ship 102 energy 148, got %d", owned.Energy)
	}
	if owned.Intimacy != 5100 {
		t.Fatalf("expected ship 102 intimacy 5100, got %d", owned.Intimacy)
	}
}
