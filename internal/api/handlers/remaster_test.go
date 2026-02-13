package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
	execTestSQL(t, "DELETE FROM remaster_progresses WHERE commander_id = $1", int64(commanderID))
	execTestSQL(t, "DELETE FROM remaster_states WHERE commander_id = $1", int64(commanderID))
	execTestSQL(t, "DELETE FROM commanders WHERE commander_id = $1", int64(commanderID))
	seedCommander(t, commanderID, "Remaster Tester")

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
	updatedEntry, err := orm.GetRemasterProgress(commanderID, 1001, 1)
	if err != nil {
		t.Fatalf("get remaster progress: %v", err)
	}
	updatedEntry.Received = true
	if err := orm.UpsertRemasterProgress(updatedEntry); err != nil {
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
	updated, err := orm.GetRemasterProgress(commanderID, 1001, 1)
	if err != nil {
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
	if _, err := orm.GetRemasterProgress(commanderID, 1001, 1); err == nil {
		t.Fatalf("expected remaster progress to be deleted")
	}
}
