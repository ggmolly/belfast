package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
)

func TestPlayerStateEndpoints(t *testing.T) {
	app := newPlayerHandlerTestApp(t)
	commanderID := uint32(9200)
	if err := orm.GormDB.Exec("DELETE FROM commander_common_flags").Error; err != nil {
		t.Fatalf("clear flags: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM commander_stories").Error; err != nil {
		t.Fatalf("clear stories: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM commander_attires").Error; err != nil {
		t.Fatalf("clear attires: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM commander_living_area_covers").Error; err != nil {
		t.Fatalf("clear covers: %v", err)
	}
	if err := orm.GormDB.Unscoped().Where("commander_id = ?", commanderID).Delete(&orm.Commander{}).Error; err != nil {
		t.Fatalf("clear commander: %v", err)
	}
	commander := orm.Commander{
		CommanderID: commanderID,
		AccountID:   1,
		Level:       1,
		Exp:         0,
		Name:        "State Tester",
		LastLogin:   time.Now().UTC(),
	}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/players/9200/flags", strings.NewReader(`{"flag_id":1000001}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/players/9200/flags", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}
	var flagsResponse struct {
		OK   bool `json:"ok"`
		Data struct {
			Flags []uint32 `json:"flags"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&flagsResponse); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(flagsResponse.Data.Flags) != 1 {
		t.Fatalf("expected 1 flag")
	}

	request = httptest.NewRequest(http.MethodPatch, "/api/v1/players/9200/guide", strings.NewReader(`{"guide_index":15,"new_guide_index":3}`))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/players/9200/guide", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}
	var guideResponse struct {
		OK   bool `json:"ok"`
		Data struct {
			GuideIndex    uint32 `json:"guide_index"`
			NewGuideIndex uint32 `json:"new_guide_index"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&guideResponse); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if guideResponse.Data.GuideIndex != 15 || guideResponse.Data.NewGuideIndex != 3 {
		t.Fatalf("unexpected guide response")
	}

	request = httptest.NewRequest(http.MethodPatch, "/api/v1/players/9200/random-flagship", strings.NewReader(`{"enabled":true}`))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/players/9200/random-flagship", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}
	var randomFlagResponse struct {
		OK   bool `json:"ok"`
		Data struct {
			Enabled bool `json:"enabled"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&randomFlagResponse); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !randomFlagResponse.Data.Enabled {
		t.Fatalf("expected random flagship enabled")
	}

	request = httptest.NewRequest(http.MethodPatch, "/api/v1/players/9200/random-flagship-mode", strings.NewReader(`{"mode":2}`))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/players/9200/random-flagship-mode", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	var randomModeResponse struct {
		OK   bool `json:"ok"`
		Data struct {
			Mode uint32 `json:"mode"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&randomModeResponse); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if randomModeResponse.Data.Mode != 2 {
		t.Fatalf("expected random flagship mode 2")
	}
	var flagEntry orm.CommanderCommonFlag
	if err := orm.GormDB.First(&flagEntry, "commander_id = ? AND flag_id = ?", commanderID, 1000007).Error; err != nil {
		t.Fatalf("expected random flagship mode flag to be set")
	}

	request = httptest.NewRequest(http.MethodPost, "/api/v1/players/9200/stories", strings.NewReader(`{"story_id":1234}`))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	request = httptest.NewRequest(http.MethodGet, "/api/v1/players/9200/stories", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	var storyResponse struct {
		OK   bool `json:"ok"`
		Data struct {
			Stories []uint32 `json:"stories"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&storyResponse); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(storyResponse.Data.Stories) != 1 {
		t.Fatalf("expected 1 story")
	}

	request = httptest.NewRequest(http.MethodPost, "/api/v1/players/9200/attires", strings.NewReader(`{"type":2,"attire_id":101}`))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	request = httptest.NewRequest(http.MethodPatch, "/api/v1/players/9200/attires/selected", strings.NewReader(`{"icon_frame_id":101}`))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	request = httptest.NewRequest(http.MethodGet, "/api/v1/players/9200/attires", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	var attireResponse struct {
		OK   bool `json:"ok"`
		Data struct {
			Attires []struct {
				Type     uint32 `json:"type"`
				AttireID uint32 `json:"attire_id"`
			} `json:"attires"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&attireResponse); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(attireResponse.Data.Attires) != 1 {
		t.Fatalf("expected 1 attire")
	}

	request = httptest.NewRequest(http.MethodPost, "/api/v1/players/9200/livingarea-covers", strings.NewReader(`{"cover_id":55}`))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	request = httptest.NewRequest(http.MethodPatch, "/api/v1/players/9200/livingarea-covers/selected", strings.NewReader(`{"cover_id":55}`))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	request = httptest.NewRequest(http.MethodGet, "/api/v1/players/9200/livingarea-covers", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	var coverResponse struct {
		OK   bool `json:"ok"`
		Data struct {
			Selected uint32   `json:"selected"`
			Owned    []uint32 `json:"owned"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&coverResponse); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if coverResponse.Data.Selected != 55 {
		t.Fatalf("expected selected cover 55")
	}
}
