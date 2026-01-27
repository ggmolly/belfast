package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/orm"
)

type chapterStateResponse struct {
	OK   bool                             `json:"ok"`
	Data types.PlayerChapterStateResponse `json:"data"`
}

type chapterStateListResponse struct {
	OK   bool                                 `json:"ok"`
	Data types.PlayerChapterStateListResponse `json:"data"`
}

func TestPlayerChapterStateEndpoints(t *testing.T) {
	app := newPlayerHandlerTestApp(t)
	commanderID := uint32(9400)
	if err := orm.GormDB.Where("commander_id = ?", commanderID).Delete(&orm.ChapterState{}).Error; err != nil {
		t.Fatalf("clear chapter state: %v", err)
	}
	if err := orm.GormDB.Unscoped().Where("commander_id = ?", commanderID).Delete(&orm.Commander{}).Error; err != nil {
		t.Fatalf("clear commander: %v", err)
	}
	commander := orm.Commander{
		CommanderID: commanderID,
		AccountID:   1,
		Level:       1,
		Exp:         0,
		Name:        "Chapter State Tester",
		LastLogin:   time.Now().UTC(),
	}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}
	state := buildChapterStatePayload()
	createPayload, err := json.Marshal(types.PlayerChapterStateCreateRequest{State: state})
	if err != nil {
		t.Fatalf("marshal create payload: %v", err)
	}
	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/players/9400/chapter-state", bytes.NewReader(createPayload))
	createRequest.Header.Set("Content-Type", "application/json")
	createResponse := httptest.NewRecorder()
	app.ServeHTTP(createResponse, createRequest)
	if createResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", createResponse.Code)
	}
	var created chapterStateResponse
	if err := json.Unmarshal(createResponse.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if !created.OK || created.Data.State.ID != 101 {
		t.Fatalf("unexpected create response: %+v", created)
	}
	getRequest := httptest.NewRequest(http.MethodGet, "/api/v1/players/9400/chapter-state", nil)
	getResponse := httptest.NewRecorder()
	app.ServeHTTP(getResponse, getRequest)
	if getResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", getResponse.Code)
	}
	var fetched chapterStateResponse
	if err := json.Unmarshal(getResponse.Body.Bytes(), &fetched); err != nil {
		t.Fatalf("decode get response: %v", err)
	}
	if !fetched.OK || fetched.Data.State.Time != state.Time {
		t.Fatalf("unexpected get response: %+v", fetched)
	}
	state.Time = 200
	updatePayload, err := json.Marshal(types.PlayerChapterStateUpdateRequest{State: state})
	if err != nil {
		t.Fatalf("marshal update payload: %v", err)
	}
	updateRequest := httptest.NewRequest(http.MethodPatch, "/api/v1/players/9400/chapter-state", bytes.NewReader(updatePayload))
	updateRequest.Header.Set("Content-Type", "application/json")
	updateResponse := httptest.NewRecorder()
	app.ServeHTTP(updateResponse, updateRequest)
	if updateResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", updateResponse.Code)
	}
	var updated chapterStateResponse
	if err := json.Unmarshal(updateResponse.Body.Bytes(), &updated); err != nil {
		t.Fatalf("decode update response: %v", err)
	}
	if updated.Data.State.Time != 200 {
		t.Fatalf("expected updated time 200")
	}
	searchRequest := httptest.NewRequest(http.MethodGet, "/api/v1/players/9400/chapter-state/search?chapter_id=101", nil)
	searchResponse := httptest.NewRecorder()
	app.ServeHTTP(searchResponse, searchRequest)
	if searchResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", searchResponse.Code)
	}
	var list chapterStateListResponse
	if err := json.Unmarshal(searchResponse.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode search response: %v", err)
	}
	if !list.OK || len(list.Data.States) != 1 {
		t.Fatalf("unexpected search response: %+v", list)
	}
	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/v1/players/9400/chapter-state", nil)
	deleteResponse := httptest.NewRecorder()
	app.ServeHTTP(deleteResponse, deleteRequest)
	if deleteResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", deleteResponse.Code)
	}
	if err := orm.GormDB.First(&orm.ChapterState{}, "commander_id = ?", commanderID).Error; err == nil {
		t.Fatalf("expected chapter state to be deleted")
	}
}

func buildChapterStatePayload() types.ChapterState {
	itemID := uint32(100)
	flag := uint32(1)
	data := uint32(0)
	return types.ChapterState{
		ID:   101,
		Time: 100,
		CellList: []types.ChapterCellInfo{
			{
				Pos:      types.ChapterCellPos{Row: 1, Column: 1},
				ItemType: 1,
				ItemID:   &itemID,
				ItemFlag: &flag,
				ItemData: &data,
				ExtraID:  []uint32{},
			},
		},
		MainGroupList: []types.ChapterGroup{
			{
				ID:               1,
				ShipList:         []types.ChapterShip{{ID: 101, HpRant: 10000}},
				Pos:              types.ChapterCellPos{Row: 1, Column: 1},
				StepCount:        0,
				BoxStrategyList:  []types.ChapterStrategy{},
				ShipStrategyList: []types.ChapterStrategy{},
				StrategyIds:      []uint32{},
				Bullet:           5,
				StartPos:         types.ChapterCellPos{Row: 1, Column: 1},
				CommanderList:    []types.ChapterCommander{},
				MoveStepDown:     0,
				KillCount:        0,
				FleetId:          1,
				VisionLv:         0,
			},
		},
		AiList:                []types.ChapterCellInfo{},
		EscortList:            []types.ChapterCellInfo{},
		Round:                 0,
		IsSubmarineAutoAttack: 0,
		OperationBuff:         []uint32{},
		ModelActCount:         0,
		BuffList:              []uint32{},
		LoopFlag:              0,
		ExtraFlagList:         []uint32{},
		CellFlagList:          []types.ChapterCellFlag{},
		ChapterHp:             0,
		ChapterStrategyList:   []types.ChapterStrategy{},
		KillCount:             0,
		InitShipCount:         1,
		ContinuousKillCount:   0,
		BattleStatistics:      []types.ChapterStrategy{},
		FleetDuties:           []types.ChapterFleetDuty{},
		MoveStepCount:         0,
		SubmarineGroupList:    []types.ChapterGroup{},
		SupportGroupList:      []types.ChapterGroup{},
	}
}
