package tests

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

type medalShopAPIResponse struct {
	OK    bool                    `json:"ok"`
	Data  types.MedalShopResponse `json:"data"`
	Error *types.APIError         `json:"error,omitempty"`
}

func resetMedalShopAPIData(t *testing.T) {
	if err := orm.GormDB.Exec("DELETE FROM medal_shop_goods").Error; err != nil {
		t.Fatalf("failed to clear medal_shop_goods: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM medal_shop_states").Error; err != nil {
		t.Fatalf("failed to clear medal_shop_states: %v", err)
	}
	if err := orm.GormDB.Where("category = ?", monthShopConfigCategory).Delete(&orm.ConfigEntry{}).Error; err != nil {
		t.Fatalf("failed to clear month shop config: %v", err)
	}
	if err := orm.GormDB.Where("category = ?", shopTemplateCategory).Delete(&orm.ConfigEntry{}).Error; err != nil {
		t.Fatalf("failed to clear shop template config: %v", err)
	}
}

func TestMedalShopAPIGet(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)
	resetMedalShopAPIData(t)
	seedMedalShopConfig(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players/1/medal-shop", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	var payload medalShopAPIResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !payload.OK {
		t.Fatalf("expected ok response")
	}
	if len(payload.Data.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(payload.Data.Items))
	}
	if payload.Data.State.NextRefreshTime <= uint32(time.Now().Unix()) {
		t.Fatalf("expected next_refresh_time in the future")
	}
}

func TestMedalShopAPIRefresh(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)
	resetMedalShopAPIData(t)
	seedMedalShopConfig(t)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/players/1/medal-shop/refresh", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	var payload medalShopAPIResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if payload.Data.State.NextRefreshTime <= uint32(time.Now().Unix()) {
		t.Fatalf("expected next_refresh_time in the future")
	}
}

func TestMedalShopAPIUpdate(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)
	resetMedalShopAPIData(t)
	seedMedalShopConfig(t)

	requestBody := []byte(`{"next_refresh_time":12345}`)
	request := httptest.NewRequest(http.MethodPut, "/api/v1/players/1/medal-shop", bytes.NewBuffer(requestBody))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	var payload medalShopAPIResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if payload.Data.State.NextRefreshTime != 12345 {
		t.Fatalf("expected next_refresh_time 12345, got %d", payload.Data.State.NextRefreshTime)
	}
}
