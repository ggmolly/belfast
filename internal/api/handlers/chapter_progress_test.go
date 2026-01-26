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

type chapterProgressResponse struct {
	OK   bool                                `json:"ok"`
	Data types.PlayerChapterProgressResponse `json:"data"`
}

type chapterProgressListResponse struct {
	OK   bool                                    `json:"ok"`
	Data types.PlayerChapterProgressListResponse `json:"data"`
}

func TestPlayerChapterProgressEndpoints(t *testing.T) {
	app := newPlayerHandlerTestApp(t)
	commanderID := uint32(9450)
	if err := orm.GormDB.Where("commander_id = ?", commanderID).Delete(&orm.ChapterProgress{}).Error; err != nil {
		t.Fatalf("clear chapter progress: %v", err)
	}
	if err := orm.GormDB.Unscoped().Where("commander_id = ?", commanderID).Delete(&orm.Commander{}).Error; err != nil {
		t.Fatalf("clear commander: %v", err)
	}
	commander := orm.Commander{
		CommanderID: commanderID,
		AccountID:   1,
		Level:       1,
		Exp:         0,
		Name:        "Chapter Progress Tester",
		LastLogin:   time.Now().UTC(),
	}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}
	progress := buildChapterProgressPayload()
	createPayload, err := json.Marshal(types.PlayerChapterProgressCreateRequest{Progress: progress})
	if err != nil {
		t.Fatalf("marshal create payload: %v", err)
	}
	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/players/9450/chapter-progress", bytes.NewReader(createPayload))
	createRequest.Header.Set("Content-Type", "application/json")
	createResponse := httptest.NewRecorder()
	app.ServeHTTP(createResponse, createRequest)
	if createResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", createResponse.Code)
	}
	var created chapterProgressResponse
	if err := json.Unmarshal(createResponse.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if !created.OK || created.Data.Progress.ChapterID != 101 {
		t.Fatalf("unexpected create response: %+v", created)
	}

	getRequest := httptest.NewRequest(http.MethodGet, "/api/v1/players/9450/chapter-progress/101", nil)
	getResponse := httptest.NewRecorder()
	app.ServeHTTP(getResponse, getRequest)
	if getResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", getResponse.Code)
	}
	var fetched chapterProgressResponse
	if err := json.Unmarshal(getResponse.Body.Bytes(), &fetched); err != nil {
		t.Fatalf("decode get response: %v", err)
	}
	if fetched.Data.Progress.Progress != 50 {
		t.Fatalf("expected progress 50, got %d", fetched.Data.Progress.Progress)
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/api/v1/players/9450/chapter-progress?limit=10", nil)
	listResponse := httptest.NewRecorder()
	app.ServeHTTP(listResponse, listRequest)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", listResponse.Code)
	}
	var list chapterProgressListResponse
	if err := json.Unmarshal(listResponse.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if !list.OK || len(list.Data.Progress) != 1 {
		t.Fatalf("unexpected list response: %+v", list)
	}

	progress.Progress = 100
	updatePayload, err := json.Marshal(types.PlayerChapterProgressUpdateRequest{Progress: progress})
	if err != nil {
		t.Fatalf("marshal update payload: %v", err)
	}
	updateRequest := httptest.NewRequest(http.MethodPatch, "/api/v1/players/9450/chapter-progress/101", bytes.NewReader(updatePayload))
	updateRequest.Header.Set("Content-Type", "application/json")
	updateResponse := httptest.NewRecorder()
	app.ServeHTTP(updateResponse, updateRequest)
	if updateResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", updateResponse.Code)
	}
	var updated chapterProgressResponse
	if err := json.Unmarshal(updateResponse.Body.Bytes(), &updated); err != nil {
		t.Fatalf("decode update response: %v", err)
	}
	if updated.Data.Progress.Progress != 100 {
		t.Fatalf("expected progress 100")
	}

	searchRequest := httptest.NewRequest(http.MethodGet, "/api/v1/players/9450/chapter-progress/search?chapter_id=101", nil)
	searchResponse := httptest.NewRecorder()
	app.ServeHTTP(searchResponse, searchRequest)
	if searchResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", searchResponse.Code)
	}
	var search chapterProgressListResponse
	if err := json.Unmarshal(searchResponse.Body.Bytes(), &search); err != nil {
		t.Fatalf("decode search response: %v", err)
	}
	if len(search.Data.Progress) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(search.Data.Progress))
	}

	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/v1/players/9450/chapter-progress/101", nil)
	deleteResponse := httptest.NewRecorder()
	app.ServeHTTP(deleteResponse, deleteRequest)
	if deleteResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", deleteResponse.Code)
	}
	if err := orm.GormDB.First(&orm.ChapterProgress{}, "commander_id = ? AND chapter_id = ?", commanderID, 101).Error; err == nil {
		t.Fatalf("expected chapter progress to be deleted")
	}
}

func buildChapterProgressPayload() types.ChapterProgress {
	return types.ChapterProgress{
		ChapterID:        101,
		Progress:         50,
		KillBossCount:    1,
		KillEnemyCount:   2,
		TakeBoxCount:     0,
		DefeatCount:      1,
		TodayDefeatCount: 1,
		PassCount:        0,
	}
}
