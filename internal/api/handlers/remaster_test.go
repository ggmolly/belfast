package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/orm"
)

type remasterStateResponse struct {
	OK   bool                              `json:"ok"`
	Data types.PlayerRemasterStateResponse `json:"data"`
}

type remasterProgressResponse struct {
	OK   bool                                 `json:"ok"`
	Data types.PlayerRemasterProgressResponse `json:"data"`
}

func TestPlayerRemasterEndpoints(t *testing.T) {
	app := newPlayerHandlerTestApp(t)
	commanderID := uint32(9300)
	if err := orm.GormDB.Where("commander_id = ?", commanderID).Delete(&orm.RemasterProgress{}).Error; err != nil {
		t.Fatalf("clear remaster progress: %v", err)
	}
	if err := orm.GormDB.Where("commander_id = ?", commanderID).Delete(&orm.RemasterState{}).Error; err != nil {
		t.Fatalf("clear remaster state: %v", err)
	}
	if err := orm.GormDB.Unscoped().Where("commander_id = ?", commanderID).Delete(&orm.Commander{}).Error; err != nil {
		t.Fatalf("clear commander: %v", err)
	}
	commander := orm.Commander{
		CommanderID: commanderID,
		AccountID:   1,
		Level:       1,
		Exp:         0,
		Name:        "Remaster Tester",
		LastLogin:   time.Now().UTC(),
	}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}

	patchPayload := strings.NewReader("{\"ticket_count\":5,\"daily_count\":2,\"last_daily_reset_at\":\"2026-01-01T00:00:00Z\"}")
	patchRequest := httptest.NewRequest(http.MethodPatch, "/api/v1/players/9300/remaster", patchPayload)
	patchRequest.Header.Set("Content-Type", "application/json")
	patchResponse := httptest.NewRecorder()
	app.ServeHTTP(patchResponse, patchRequest)
	if patchResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", patchResponse.Code)
	}
	var stateResponse remasterStateResponse
	if err := json.Unmarshal(patchResponse.Body.Bytes(), &stateResponse); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !stateResponse.OK || stateResponse.Data.TicketCount != 5 || stateResponse.Data.DailyCount != 2 {
		t.Fatalf("unexpected remaster state response: %+v", stateResponse)
	}

	createPayload := strings.NewReader("{\"chapter_id\":1001,\"pos\":1,\"count\":3,\"received\":false}")
	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/players/9300/remaster/progress", createPayload)
	createRequest.Header.Set("Content-Type", "application/json")
	createResponse := httptest.NewRecorder()
	app.ServeHTTP(createResponse, createRequest)
	if createResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", createResponse.Code)
	}
	if err := orm.GormDB.Model(&orm.RemasterProgress{}).
		Where("commander_id = ? AND chapter_id = ? AND pos = ?", commanderID, 1001, 1).
		Update("received", true).Error; err != nil {
		t.Fatalf("update remaster received: %v", err)
	}
	countOnlyPayload := strings.NewReader("{\"chapter_id\":1001,\"pos\":1,\"count\":4}")
	countOnlyRequest := httptest.NewRequest(http.MethodPost, "/api/v1/players/9300/remaster/progress", countOnlyPayload)
	countOnlyRequest.Header.Set("Content-Type", "application/json")
	countOnlyResponse := httptest.NewRecorder()
	app.ServeHTTP(countOnlyResponse, countOnlyRequest)
	if countOnlyResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", countOnlyResponse.Code)
	}
	var updated orm.RemasterProgress
	if err := orm.GormDB.First(&updated, "commander_id = ? AND chapter_id = ? AND pos = ?", commanderID, 1001, 1).Error; err != nil {
		t.Fatalf("load remaster progress: %v", err)
	}
	if !updated.Received {
		t.Fatalf("expected received to remain true")
	}

	getRequest := httptest.NewRequest(http.MethodGet, "/api/v1/players/9300/remaster/progress?chapter_id=1001", nil)
	getResponse := httptest.NewRecorder()
	app.ServeHTTP(getResponse, getRequest)
	if getResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", getResponse.Code)
	}
	var progressResponse remasterProgressResponse
	if err := json.Unmarshal(getResponse.Body.Bytes(), &progressResponse); err != nil {
		t.Fatalf("decode progress response: %v", err)
	}
	if !progressResponse.OK || len(progressResponse.Data.Progress) != 1 {
		t.Fatalf("unexpected progress response: %+v", progressResponse)
	}
	if progressResponse.Data.Progress[0].Count != 4 {
		t.Fatalf("unexpected progress count")
	}

	updatePayload := strings.NewReader("{\"received\":true}")
	updateRequest := httptest.NewRequest(http.MethodPatch, "/api/v1/players/9300/remaster/progress/1001/1", updatePayload)
	updateRequest.Header.Set("Content-Type", "application/json")
	updateResponse := httptest.NewRecorder()
	app.ServeHTTP(updateResponse, updateRequest)
	if updateResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", updateResponse.Code)
	}

	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/v1/players/9300/remaster/progress/1001/1", nil)
	deleteResponse := httptest.NewRecorder()
	app.ServeHTTP(deleteResponse, deleteRequest)
	if deleteResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", deleteResponse.Code)
	}
	if err := orm.GormDB.First(&orm.RemasterProgress{}, "commander_id = ? AND chapter_id = ? AND pos = ?", commanderID, 1001, 1).Error; err == nil {
		t.Fatalf("expected remaster progress to be deleted")
	}
}
