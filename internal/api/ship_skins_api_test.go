package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/orm"
)

type shipSkinsResponse struct {
	OK   bool                   `json:"ok"`
	Data types.SkinListResponse `json:"data"`
}

func TestShipSkinsList(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	skinA := orm.Skin{ID: 10, Name: "Skin A", ShipGroup: 1}
	skinB := orm.Skin{ID: 11, Name: "Skin B", ShipGroup: 2}
	if err := orm.GormDB.Create(&skinA).Error; err != nil {
		t.Fatalf("failed to create skinA: %v", err)
	}
	if err := orm.GormDB.Create(&skinB).Error; err != nil {
		t.Fatalf("failed to create skinB: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/ships/1/skins?offset=0&limit=10", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload shipSkinsResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !payload.OK {
		t.Fatalf("expected ok response")
	}
	if payload.Data.Meta.Total != 1 {
		t.Fatalf("expected 1 skin, got %d", payload.Data.Meta.Total)
	}
}
