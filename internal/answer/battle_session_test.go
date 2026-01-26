package answer

import (
	"errors"
	"reflect"
	"testing"

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
