package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
)

func clearCommanders(t *testing.T) {
	t.Helper()
	execTestSQL(t, "DELETE FROM commanders")
}

func clearCommanderCommonFlags(t *testing.T) {
	t.Helper()
	execTestSQL(t, "DELETE FROM commander_common_flags")
}

func seedCommander(t *testing.T, commanderID uint32, name string) {
	t.Helper()
	if err := orm.CreateCommanderRoot(commanderID, 1, name, 0, 0); err != nil {
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

	flags, err := orm.ListCommanderCommonFlags(99999)
	if err != nil {
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

	commander, err := orm.GetCommanderCoreByID(99999)
	if err != nil {
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

	count := queryInt64TestSQL(t, "SELECT COUNT(1) FROM commanders WHERE deleted_at IS NULL")
	if count != 0 {
		t.Fatalf("expected 0 commanders, got %d", count)
	}
}

func TestPlayerBuildReturnsHydratedShipName(t *testing.T) {
	app := newPlayerHandlerTestApp(t)
	commanderID := uint32(99998)
	shipID := uint32(7001)

	execTestSQL(t, "DELETE FROM builds WHERE builder_id = $1", int64(commanderID))
	execTestSQL(t, "DELETE FROM ships WHERE template_id = $1", int64(shipID))
	execTestSQL(t, "DELETE FROM commanders WHERE commander_id = $1", int64(commanderID))
	t.Cleanup(func() {
		execTestSQL(t, "DELETE FROM builds WHERE builder_id = $1", int64(commanderID))
		execTestSQL(t, "DELETE FROM ships WHERE template_id = $1", int64(shipID))
		execTestSQL(t, "DELETE FROM commanders WHERE commander_id = $1", int64(commanderID))
	})

	seedCommander(t, commanderID, "BuildPlayer")
	execTestSQL(t, "INSERT INTO rarities (id, name) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING", int64(2), "Common")
	ship := orm.Ship{TemplateID: shipID, Name: "Hydrated Ship", EnglishName: "Hydrated Ship", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	if err := ship.Create(); err != nil {
		t.Fatalf("seed ship: %v", err)
	}

	build := orm.Build{BuilderID: commanderID, ShipID: shipID, PoolID: 1, FinishesAt: time.Now().UTC().Add(time.Hour)}
	if err := build.Create(); err != nil {
		t.Fatalf("seed build: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players/99998/builds/"+strconv.FormatUint(uint64(build.ID), 10), nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}

	var responseStruct struct {
		OK   bool `json:"ok"`
		Data struct {
			BuildID  uint32 `json:"build_id"`
			ShipName string `json:"ship_name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}
	if responseStruct.Data.BuildID != build.ID {
		t.Fatalf("expected build id %d, got %d", build.ID, responseStruct.Data.BuildID)
	}
	if responseStruct.Data.ShipName != ship.Name {
		t.Fatalf("expected ship name %q, got %q", ship.Name, responseStruct.Data.ShipName)
	}
}
