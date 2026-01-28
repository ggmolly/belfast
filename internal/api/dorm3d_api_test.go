package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/orm"
)

type dorm3dDetailResponse struct {
	OK   bool                  `json:"ok"`
	Data types.Dorm3dApartment `json:"data"`
}

type dorm3dListResponse struct {
	OK   bool                              `json:"ok"`
	Data types.Dorm3dApartmentListResponse `json:"data"`
}

func TestDorm3dApartmentCRUD(t *testing.T) {
	setupTestAPI(t)
	if err := orm.GormDB.Exec("DELETE FROM dorm3d_apartments").Error; err != nil {
		t.Fatalf("failed to clear dorm3d_apartments: %v", err)
	}
	requestPayload := types.Dorm3dApartmentRequest{
		CommanderID:        9200,
		DailyVigorMax:      45,
		Gifts:              orm.Dorm3dGiftList{{GiftID: 1, Number: 2, UsedNumber: 0}},
		GiftDaily:          orm.Dorm3dGiftShopList{},
		GiftPermanent:      orm.Dorm3dGiftShopList{},
		FurnitureDaily:     orm.Dorm3dGiftShopList{},
		FurniturePermanent: orm.Dorm3dGiftShopList{},
		Rooms:              orm.Dorm3dRoomList{},
		Ships:              orm.Dorm3dShipList{},
		Ins:                orm.Dorm3dInsList{},
	}
	body, err := json.Marshal(requestPayload)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/dorm3d-apartments", bytes.NewBuffer(body))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/dorm3d-apartments/9200", nil)
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	var detail dorm3dDetailResponse
	if err := json.NewDecoder(response.Body).Decode(&detail); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !detail.OK {
		t.Fatalf("expected ok response")
	}
	if detail.Data.DailyVigorMax != 45 {
		t.Fatalf("expected daily_vigor_max 45, got %d", detail.Data.DailyVigorMax)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/dorm3d-apartments?offset=0&limit=10", nil)
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	var listResponse dorm3dListResponse
	if err := json.NewDecoder(response.Body).Decode(&listResponse); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if listResponse.Data.Meta.Total == 0 {
		t.Fatalf("expected at least 1 apartment")
	}

	requestPayload.DailyVigorMax = 55
	body, err = json.Marshal(requestPayload)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}
	request = httptest.NewRequest(http.MethodPut, "/api/v1/dorm3d-apartments/9200", bytes.NewBuffer(body))
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/dorm3d-apartments/9200", nil)
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if err := json.NewDecoder(response.Body).Decode(&detail); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if detail.Data.DailyVigorMax != 55 {
		t.Fatalf("expected daily_vigor_max 55, got %d", detail.Data.DailyVigorMax)
	}

	request = httptest.NewRequest(http.MethodDelete, "/api/v1/dorm3d-apartments/9200", nil)
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/dorm3d-apartments/9200", nil)
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", response.Code)
	}
}
