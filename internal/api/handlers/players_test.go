package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
)

func clearCommanders(t *testing.T) {
	t.Helper()
	if err := orm.GormDB.Exec("DELETE FROM commanders").Error; err != nil {
		t.Fatalf("clear commanders: %v", err)
	}
}

func clearCommanderCommonFlags(t *testing.T) {
	t.Helper()
	if err := orm.GormDB.Exec("DELETE FROM commander_common_flags").Error; err != nil {
		t.Fatalf("clear commander common flags: %v", err)
	}
}

func seedCommander(t *testing.T, commanderID uint32, name string) {
	t.Helper()
	commander := orm.Commander{
		CommanderID: commanderID,
		Name:        name,
		Level:       1,
		Exp:         0,
	}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("seed commander: %v", err)
	}
}

func TestPlayerGetFlagsReturnsEmpty(t *testing.T) {
	app := newPlayerHandlerTestApp(t)
	clearCommanderCommonFlags(t)
	seedCommander(t, 99999, "TestPlayer")
	t.Cleanup(func() {
		clearCommanders(t)
		clearCommanderCommonFlags(t)
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players/99999/flags", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}

	contentType := response.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		t.Fatalf("expected application/json content type, got %s", contentType)
	}

	var responseStruct struct {
		OK   bool `json:"ok"`
		Data struct {
			Flags []uint32 `json:"flags"`
		} `json:"data"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}
	if len(responseStruct.Data.Flags) != 0 {
		t.Fatalf("expected empty flags list, got %d flags", len(responseStruct.Data.Flags))
	}
}

func TestPlayerPostFlag(t *testing.T) {
	app := newPlayerHandlerTestApp(t)
	clearCommanderCommonFlags(t)
	seedCommander(t, 99999, "TestPlayer")
	t.Cleanup(func() {
		clearCommanders(t)
		clearCommanderCommonFlags(t)
	})

	requestBody := `{"flag_id":1000001}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/players/99999/flags", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK   bool `json:"ok"`
		Data any  `json:"data"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}

	var flags []uint32
	if err := orm.GormDB.Table("commander_common_flags").Select("flag_id").Find(&flags).Error; err != nil {
		t.Fatalf("query flags failed: %v", err)
	}
	if len(flags) != 1 {
		t.Fatalf("expected 1 flag in db, got %d", len(flags))
	}
	if flags[0] != 1000001 {
		t.Fatalf("expected flag 1000001 in db, got %d", flags[0])
	}

	getRequest := httptest.NewRequest(http.MethodGet, "/api/v1/players/99999/flags", nil)
	getResponse := httptest.NewRecorder()
	app.ServeHTTP(getResponse, getRequest)

	if getResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200 on get, got %d", getResponse.Code)
	}

	var getResponseStruct struct {
		OK   bool `json:"ok"`
		Data struct {
			Flags []uint32 `json:"flags"`
		} `json:"data"`
	}

	if err := json.NewDecoder(getResponse.Body).Decode(&getResponseStruct); err != nil {
		t.Fatalf("decode get response failed: %v", err)
	}

	if !getResponseStruct.OK {
		t.Fatalf("expected ok true on get")
	}
	if len(getResponseStruct.Data.Flags) != 1 {
		t.Fatalf("expected 1 flag in get response, got %d", len(getResponseStruct.Data.Flags))
	}
	if getResponseStruct.Data.Flags[0] != 1000001 {
		t.Fatalf("expected flag 1000001 in get response, got %d", getResponseStruct.Data.Flags[0])
	}
}

func TestPlayerPatchGuideIndex(t *testing.T) {
	app := newPlayerHandlerTestApp(t)
	clearCommanders(t)
	seedCommander(t, 99999, "TestPlayer")
	t.Cleanup(func() {
		clearCommanders(t)
	})

	requestBody := `{"guide_index":15,"new_guide_index":3}`
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/players/99999/guide", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK   bool `json:"ok"`
		Data any  `json:"data"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}

	commander := orm.Commander{}
	if err := orm.GormDB.Where("commander_id = ?", 99999).First(&commander).Error; err != nil {
		t.Fatalf("query commander failed: %v", err)
	}
	if commander.GuideIndex != 15 {
		t.Fatalf("expected guide index 15 after patch, got %d", commander.GuideIndex)
	}
	if commander.NewGuideIndex != 3 {
		t.Fatalf("expected new guide index 3 after patch, got %d", commander.NewGuideIndex)
	}
}

func TestPlayerGetGuideIndexReturnsUpdated(t *testing.T) {
	app := newPlayerHandlerTestApp(t)
	clearCommanders(t)
	seedCommander(t, 99999, "TestPlayer")
	t.Cleanup(func() {
		clearCommanders(t)
	})

	requestBody := `{"guide_index":15,"new_guide_index":3}`
	patchRequest := httptest.NewRequest(http.MethodPatch, "/api/v1/players/99999/guide", strings.NewReader(requestBody))
	patchRequest.Header.Set("Content-Type", "application/json")
	patchResponse := httptest.NewRecorder()
	app.ServeHTTP(patchResponse, patchRequest)
	if patchResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", patchResponse.Code, patchResponse.Body.String())
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players/99999/guide", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}

	var responseStruct struct {
		OK   bool `json:"ok"`
		Data struct {
			GuideIndex    uint32 `json:"guide_index"`
			NewGuideIndex uint32 `json:"new_guide_index"`
		} `json:"data"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}
	if responseStruct.Data.GuideIndex == 0 {
		t.Fatalf("expected guide index to be set")
	}
	if responseStruct.Data.NewGuideIndex == 0 {
		t.Fatalf("expected new guide index to be set")
	}
}

func TestPlayerDeleteNotFound(t *testing.T) {
	app := newPlayerHandlerTestApp(t)
	clearCommanders(t)

	request := httptest.NewRequest(http.MethodDelete, "/api/v1/players/99999", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", response.Code)
	}

	var responseStruct struct {
		OK bool `json:"ok"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if responseStruct.OK {
		t.Fatalf("expected ok false for not found")
	}

	var count int64
	if err := orm.GormDB.Table("commanders").Count(&count).Error; err != nil {
		t.Fatalf("count commanders failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 commanders, got %d", count)
	}
}
