package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
	execTestSQL(t, "DELETE FROM chapter_states WHERE commander_id = $1", int64(commanderID))
	execTestSQL(t, "DELETE FROM commanders WHERE commander_id = $1", int64(commanderID))
	seedCommander(t, commanderID, "Chapter State Tester")
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
	if _, err := orm.GetChapterState(commanderID); err == nil {
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

func TestPlayerChapterStateSearchUsesDBPaginationMeta(t *testing.T) {
	app := newPlayerHandlerTestApp(t)
	commanderID := uint32(9401)
	execTestSQL(t, "DELETE FROM chapter_states WHERE commander_id = $1", int64(commanderID))
	execTestSQL(t, "DELETE FROM commanders WHERE commander_id = $1", int64(commanderID))
	seedCommander(t, commanderID, "Chapter State Pagination Tester")

	state := buildChapterStatePayload()
	createPayload, err := json.Marshal(types.PlayerChapterStateCreateRequest{State: state})
	if err != nil {
		t.Fatalf("marshal create payload: %v", err)
	}
	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/players/9401/chapter-state", bytes.NewReader(createPayload))
	createRequest.Header.Set("Content-Type", "application/json")
	createResponse := httptest.NewRecorder()
	app.ServeHTTP(createResponse, createRequest)
	if createResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", createResponse.Code)
	}

	execTestSQL(t, "UPDATE chapter_states SET updated_at = $2 WHERE commander_id = $1", int64(commanderID), int64(500))

	searchRequest := httptest.NewRequest(http.MethodGet, "/api/v1/players/9401/chapter-state/search?offset=1&limit=1", nil)
	searchResponse := httptest.NewRecorder()
	app.ServeHTTP(searchResponse, searchRequest)
	if searchResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", searchResponse.Code)
	}
	var paged chapterStateListResponse
	if err := json.Unmarshal(searchResponse.Body.Bytes(), &paged); err != nil {
		t.Fatalf("decode paged search response: %v", err)
	}
	if paged.Data.Meta.Total != 1 {
		t.Fatalf("expected total 1, got %d", paged.Data.Meta.Total)
	}
	if len(paged.Data.States) != 0 {
		t.Fatalf("expected empty page, got %d", len(paged.Data.States))
	}

	filteredRequest := httptest.NewRequest(http.MethodGet, "/api/v1/players/9401/chapter-state/search?updated_since=1970-01-01T00:08:20Z", nil)
	filteredResponse := httptest.NewRecorder()
	app.ServeHTTP(filteredResponse, filteredRequest)
	if filteredResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", filteredResponse.Code)
	}
	var filtered chapterStateListResponse
	if err := json.Unmarshal(filteredResponse.Body.Bytes(), &filtered); err != nil {
		t.Fatalf("decode filtered search response: %v", err)
	}
	if filtered.Data.Meta.Total != 1 || len(filtered.Data.States) != 1 {
		t.Fatalf("unexpected filtered response: %+v", filtered)
	}
}
