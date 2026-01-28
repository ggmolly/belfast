package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/orm"
)

type shipListResponse struct {
	OK   bool                   `json:"ok"`
	Data types.ShipListResponse `json:"data"`
}

type shipDetailResponse struct {
	OK   bool              `json:"ok"`
	Data types.ShipSummary `json:"data"`
}

type itemListResponse struct {
	OK   bool                   `json:"ok"`
	Data types.ItemListResponse `json:"data"`
}

type itemDetailResponse struct {
	OK   bool              `json:"ok"`
	Data types.ItemSummary `json:"data"`
}

type resourceListResponse struct {
	OK   bool                       `json:"ok"`
	Data types.ResourceListResponse `json:"data"`
}

type resourceDetailResponse struct {
	OK   bool                  `json:"ok"`
	Data types.ResourceSummary `json:"data"`
}

type skinListResponse struct {
	OK   bool                   `json:"ok"`
	Data types.SkinListResponse `json:"data"`
}

type skinDetailResponse struct {
	OK   bool              `json:"ok"`
	Data types.SkinSummary `json:"data"`
}

func TestShipListFilters(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	shipA := orm.Ship{TemplateID: 2, Name: "Alpha", RarityID: 3, Star: 1, Type: 1, Nationality: 1}
	shipB := orm.Ship{TemplateID: 3, Name: "Bravo", RarityID: 4, Star: 1, Type: 2, Nationality: 2}
	if err := orm.GormDB.Create(&shipA).Error; err != nil {
		t.Fatalf("failed to create shipA: %v", err)
	}
	if err := orm.GormDB.Create(&shipB).Error; err != nil {
		t.Fatalf("failed to create shipB: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/ships?rarity=3&name=alpha&offset=0&limit=10", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload shipListResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !payload.OK {
		t.Fatalf("expected ok response")
	}
	if len(payload.Data.Ships) != 1 {
		t.Fatalf("expected 1 ship, got %d", len(payload.Data.Ships))
	}
	if payload.Data.Ships[0].ID != 2 {
		t.Fatalf("expected ship id 2, got %d", payload.Data.Ships[0].ID)
	}
}

func TestShipDetail(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/ships/1", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload shipDetailResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !payload.OK {
		t.Fatalf("expected ok response")
	}
	if payload.Data.ID != 1 {
		t.Fatalf("expected ship id 1, got %d", payload.Data.ID)
	}
}

func TestItemList(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/items?offset=0&limit=10", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload itemListResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !payload.OK {
		t.Fatalf("expected ok response")
	}
	if payload.Data.Meta.Total < 1 {
		t.Fatalf("expected at least 1 item, got %d", payload.Data.Meta.Total)
	}
}

func TestItemListWithoutLimit(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/items?offset=0", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload itemListResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !payload.OK {
		t.Fatalf("expected ok response")
	}
	if payload.Data.Meta.Total != int64(len(payload.Data.Items)) {
		t.Fatalf("expected all items, got %d of %d", len(payload.Data.Items), payload.Data.Meta.Total)
	}
}

func TestItemDetail(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/items/20001", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload itemDetailResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !payload.OK {
		t.Fatalf("expected ok response")
	}
	if payload.Data.ID != 20001 {
		t.Fatalf("expected item id 20001, got %d", payload.Data.ID)
	}
}

func TestResourceList(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/resources?offset=0&limit=10", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload resourceListResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !payload.OK {
		t.Fatalf("expected ok response")
	}
	if payload.Data.Meta.Total < 1 {
		t.Fatalf("expected at least 1 resource, got %d", payload.Data.Meta.Total)
	}
}

func TestResourceDetail(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/resources/1", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload resourceDetailResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !payload.OK {
		t.Fatalf("expected ok response")
	}
	if payload.Data.ID != 1 {
		t.Fatalf("expected resource id 1, got %d", payload.Data.ID)
	}
}

func TestSkinList(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	skinA := orm.Skin{ID: 1, Name: "Skin A", ShipGroup: 1}
	skinB := orm.Skin{ID: 2, Name: "Skin B", ShipGroup: 2}
	if err := orm.GormDB.Create(&skinA).Error; err != nil {
		t.Fatalf("failed to create skinA: %v", err)
	}
	if err := orm.GormDB.Create(&skinB).Error; err != nil {
		t.Fatalf("failed to create skinB: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/skins?offset=0&limit=10", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload skinListResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !payload.OK {
		t.Fatalf("expected ok response")
	}
	if payload.Data.Meta.Total != 2 {
		t.Fatalf("expected 2 skins, got %d", payload.Data.Meta.Total)
	}
}

func TestSkinDetail(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	skin := orm.Skin{ID: 1, Name: "Skin"}
	if err := orm.GormDB.Create(&skin).Error; err != nil {
		t.Fatalf("failed to create skin: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/skins/1", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload skinDetailResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !payload.OK {
		t.Fatalf("expected ok response")
	}
	if payload.Data.ID != 1 {
		t.Fatalf("expected skin id 1, got %d", payload.Data.ID)
	}
}
