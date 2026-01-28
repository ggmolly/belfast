package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/orm"
)

func newDorm3dTestApp(t *testing.T) *iris.Application {
	initPlayerHandlerTestDB(t)
	app := iris.New()
	handler := NewDorm3dHandler()
	RegisterDorm3dRoutes(app.Party("/api/v1/dorm3d-apartments"), handler)
	if err := app.Build(); err != nil {
		t.Fatalf("build app: %v", err)
	}
	return app
}

func clearDorm3dApartments(t *testing.T) {
	t.Helper()
	if err := orm.GormDB.Exec("DELETE FROM dorm3d_apartments").Error; err != nil {
		t.Fatalf("clear dorm3d apartments: %v", err)
	}
}

func seedDorm3dApartment(t *testing.T, commanderID uint32, dailyVigorMax uint32) {
	t.Helper()
	apartment := orm.Dorm3dApartment{
		CommanderID:        commanderID,
		DailyVigorMax:      dailyVigorMax,
		Gifts:              orm.Dorm3dGiftList{},
		Ships:              orm.Dorm3dShipList{},
		GiftDaily:          orm.Dorm3dGiftShopList{},
		GiftPermanent:      orm.Dorm3dGiftShopList{},
		FurnitureDaily:     orm.Dorm3dGiftShopList{},
		FurniturePermanent: orm.Dorm3dGiftShopList{},
		Rooms:              orm.Dorm3dRoomList{},
		Ins:                orm.Dorm3dInsList{},
	}
	if err := orm.GormDB.Create(&apartment).Error; err != nil {
		t.Fatalf("seed dorm3d apartment: %v", err)
	}
}

func TestDorm3dListReturnsEmpty(t *testing.T) {
	app := newDorm3dTestApp(t)
	clearDorm3dApartments(t)
	t.Cleanup(func() {
		clearDorm3dApartments(t)
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/dorm3d-apartments", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK   bool `json:"ok"`
		Data struct {
			Apartments []orm.Dorm3dApartment `json:"apartments"`
			Meta       struct {
				Offset uint32 `json:"offset"`
				Limit  uint32 `json:"limit"`
				Total  int64  `json:"total"`
			} `json:"meta"`
		} `json:"data"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}
	if len(responseStruct.Data.Apartments) != 0 {
		t.Fatalf("expected empty apartments list, got %d", len(responseStruct.Data.Apartments))
	}
	if responseStruct.Data.Meta.Total != 0 {
		t.Fatalf("expected total 0, got %d", responseStruct.Data.Meta.Total)
	}
}

func TestDorm3dCreateApartment(t *testing.T) {
	app := newDorm3dTestApp(t)
	clearDorm3dApartments(t)
	t.Cleanup(func() {
		clearDorm3dApartments(t)
	})

	requestBody := `{"commander_id":99998,"daily_vigor_max":100}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/dorm3d-apartments", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK bool `json:"ok"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}

	var apartment orm.Dorm3dApartment
	if err := orm.GormDB.Where("commander_id = ?", 99998).First(&apartment).Error; err != nil {
		t.Fatalf("query apartment failed: %v", err)
	}
	if apartment.CommanderID != 99998 {
		t.Fatalf("expected commander_id 99998, got %d", apartment.CommanderID)
	}
	if apartment.DailyVigorMax != 100 {
		t.Fatalf("expected daily_vigor_max 100, got %d", apartment.DailyVigorMax)
	}
}

func TestDorm3dCreateMissingCommanderID(t *testing.T) {
	app := newDorm3dTestApp(t)
	clearDorm3dApartments(t)

	requestBody := `{"daily_vigor_max":100}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/dorm3d-apartments", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}

	var responseStruct struct {
		OK    bool `json:"ok"`
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if responseStruct.OK {
		t.Fatalf("expected ok false for missing commander_id")
	}
}

func TestDorm3dGetApartment(t *testing.T) {
	app := newDorm3dTestApp(t)
	clearDorm3dApartments(t)
	seedDorm3dApartment(t, 99999, 200)
	t.Cleanup(func() {
		clearDorm3dApartments(t)
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/dorm3d-apartments/99999", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK   bool                `json:"ok"`
		Data orm.Dorm3dApartment `json:"data"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}
	if responseStruct.Data.CommanderID != 99999 {
		t.Fatalf("expected commander_id 99999, got %d", responseStruct.Data.CommanderID)
	}
	if responseStruct.Data.DailyVigorMax != 200 {
		t.Fatalf("expected daily_vigor_max 200, got %d", responseStruct.Data.DailyVigorMax)
	}
}

func TestDorm3dGetNotFound(t *testing.T) {
	app := newDorm3dTestApp(t)
	clearDorm3dApartments(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/dorm3d-apartments/99999", nil)
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
}

func TestDorm3dUpdateApartment(t *testing.T) {
	app := newDorm3dTestApp(t)
	clearDorm3dApartments(t)
	seedDorm3dApartment(t, 99999, 100)
	t.Cleanup(func() {
		clearDorm3dApartments(t)
	})

	requestBody := `{"daily_vigor_max":250}`
	request := httptest.NewRequest(http.MethodPut, "/api/v1/dorm3d-apartments/99999", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK bool `json:"ok"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}

	var apartment orm.Dorm3dApartment
	if err := orm.GormDB.Where("commander_id = ?", 99999).First(&apartment).Error; err != nil {
		t.Fatalf("query apartment failed: %v", err)
	}
	if apartment.DailyVigorMax != 250 {
		t.Fatalf("expected daily_vigor_max 250 after update, got %d", apartment.DailyVigorMax)
	}
}

func TestDorm3dUpdateNotFound(t *testing.T) {
	app := newDorm3dTestApp(t)
	clearDorm3dApartments(t)

	requestBody := `{"daily_vigor_max":250}`
	request := httptest.NewRequest(http.MethodPut, "/api/v1/dorm3d-apartments/99999", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
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
}

func TestDorm3dDeleteApartment(t *testing.T) {
	app := newDorm3dTestApp(t)
	clearDorm3dApartments(t)
	seedDorm3dApartment(t, 99999, 100)
	t.Cleanup(func() {
		clearDorm3dApartments(t)
	})

	request := httptest.NewRequest(http.MethodDelete, "/api/v1/dorm3d-apartments/99999", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK bool `json:"ok"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}

	var apartment orm.Dorm3dApartment
	if err := orm.GormDB.Where("commander_id = ?", 99999).First(&apartment).Error; err == nil {
		t.Fatalf("expected apartment to be deleted")
	}
}

func TestDorm3dDeleteNotFound(t *testing.T) {
	app := newDorm3dTestApp(t)
	clearDorm3dApartments(t)

	request := httptest.NewRequest(http.MethodDelete, "/api/v1/dorm3d-apartments/99999", nil)
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
}
