package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ggmolly/belfast/internal/api/types"
)

type loveLetterStateResponse struct {
	OK   bool                                `json:"ok"`
	Data types.PlayerLoveLetterStateResponse `json:"data"`
}

func TestPlayerLoveLetterStateEndpoints(t *testing.T) {
	app := newPlayerHandlerTestApp(t)
	commanderID := uint32(9350)
	execTestSQL(t, "DELETE FROM commander_love_letter_states WHERE commander_id = $1", int64(commanderID))
	execTestSQL(t, "DELETE FROM commanders WHERE commander_id = $1", int64(commanderID))
	seedCommander(t, commanderID, "Love Letter Tester")

	patchPayload := strings.NewReader(`{
		"medals":[{"group_id":10000,"exp":10,"level":1}],
		"manual_letters":[{"group_id":10000,"letter_id_list":[2018001]}],
		"converted_items":[{"item_id":41002,"group_id":10000,"year":2018}],
		"rewarded_ids":[1,2],
		"letter_contents":{"2018001":"dear commander"}
	}`)
	patchRequest := httptest.NewRequest(http.MethodPatch, "/api/v1/players/9350/love-letter", patchPayload)
	patchRequest.Header.Set("Content-Type", "application/json")
	patchResponse := httptest.NewRecorder()
	app.ServeHTTP(patchResponse, patchRequest)
	if patchResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", patchResponse.Code)
	}
	var stateResponse loveLetterStateResponse
	if err := json.Unmarshal(patchResponse.Body.Bytes(), &stateResponse); err != nil {
		t.Fatalf("decode patch response: %v", err)
	}
	if !stateResponse.OK || len(stateResponse.Data.Medals) != 1 || len(stateResponse.Data.ConvertedItems) != 1 {
		t.Fatalf("unexpected patch payload: %+v", stateResponse)
	}
	if stateResponse.Data.LetterContents[2018001] != "dear commander" {
		t.Fatalf("unexpected letter content map: %+v", stateResponse.Data.LetterContents)
	}

	getRequest := httptest.NewRequest(http.MethodGet, "/api/v1/players/9350/love-letter", nil)
	getResponse := httptest.NewRecorder()
	app.ServeHTTP(getResponse, getRequest)
	if getResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", getResponse.Code)
	}
	stateResponse = loveLetterStateResponse{}
	if err := json.Unmarshal(getResponse.Body.Bytes(), &stateResponse); err != nil {
		t.Fatalf("decode get response: %v", err)
	}
	if !stateResponse.OK || len(stateResponse.Data.ManualLetters) != 1 || len(stateResponse.Data.RewardedIDs) != 2 {
		t.Fatalf("unexpected get payload: %+v", stateResponse)
	}

	badPatchRequest := httptest.NewRequest(http.MethodPatch, "/api/v1/players/9350/love-letter", strings.NewReader(`{}`))
	badPatchRequest.Header.Set("Content-Type", "application/json")
	badPatchResponse := httptest.NewRecorder()
	app.ServeHTTP(badPatchResponse, badPatchRequest)
	if badPatchResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for empty patch, got %d", badPatchResponse.Code)
	}

	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/v1/players/9350/love-letter", nil)
	deleteResponse := httptest.NewRecorder()
	app.ServeHTTP(deleteResponse, deleteRequest)
	if deleteResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", deleteResponse.Code)
	}

	getAfterDeleteRequest := httptest.NewRequest(http.MethodGet, "/api/v1/players/9350/love-letter", nil)
	getAfterDeleteResponse := httptest.NewRecorder()
	app.ServeHTTP(getAfterDeleteResponse, getAfterDeleteRequest)
	if getAfterDeleteResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", getAfterDeleteResponse.Code)
	}
	stateResponse = loveLetterStateResponse{}
	if err := json.Unmarshal(getAfterDeleteResponse.Body.Bytes(), &stateResponse); err != nil {
		t.Fatalf("decode get-after-delete response: %v", err)
	}
	if len(stateResponse.Data.Medals) != 0 || len(stateResponse.Data.RewardedIDs) != 0 {
		t.Fatalf("expected empty state after delete, got %+v", stateResponse.Data)
	}
}
