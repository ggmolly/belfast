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

func newExchangeCodeTestApp(t *testing.T) *iris.Application {
	initPlayerHandlerTestDB(t)
	app := iris.New()
	handler := NewExchangeCodeHandler()
	RegisterExchangeCodeRoutes(app.Party("/api/v1/exchange-codes"), handler)
	if err := app.Build(); err != nil {
		t.Fatalf("build app: %v", err)
	}
	return app
}

func clearExchangeCodes(t *testing.T) {
	t.Helper()
	if err := orm.GormDB.Exec("DELETE FROM exchange_codes").Error; err != nil {
		t.Fatalf("clear exchange codes: %v", err)
	}
}

func seedExchangeCode(t *testing.T, id uint32, code string, platform string, quota int) {
	t.Helper()
	rewards, _ := json.Marshal([]map[string]interface{}{
		{"type": 1, "id": 1001, "count": 100},
	})
	exchangeCode := orm.ExchangeCode{
		ID:       id,
		Code:     code,
		Platform: platform,
		Quota:    quota,
		Rewards:  rewards,
	}
	if err := orm.GormDB.Create(&exchangeCode).Error; err != nil {
		t.Fatalf("seed exchange code: %v", err)
	}
}

func TestListExchangeCodesReturnsEmpty(t *testing.T) {
	app := newExchangeCodeTestApp(t)
	clearExchangeCodes(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/exchange-codes", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK   bool `json:"ok"`
		Data struct {
			Codes []struct {
				ID       uint32 `json:"id"`
				Code     string `json:"code"`
				Platform string `json:"platform"`
				Quota    int    `json:"quota"`
				Rewards  []struct {
					Type  uint32 `json:"type"`
					ID    uint32 `json:"id"`
					Count uint32 `json:"count"`
				} `json:"rewards"`
			} `json:"codes"`
			Meta struct {
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
	if len(responseStruct.Data.Codes) != 0 {
		t.Fatalf("expected empty codes list, got %d", len(responseStruct.Data.Codes))
	}
	if responseStruct.Data.Meta.Total != 0 {
		t.Fatalf("expected total 0, got %d", responseStruct.Data.Meta.Total)
	}
}

func TestListExchangeCodesReturnsData(t *testing.T) {
	app := newExchangeCodeTestApp(t)
	clearExchangeCodes(t)
	seedExchangeCode(t, 1, "TESTCODE1", "all", 100)
	seedExchangeCode(t, 2, "TESTCODE2", "android", 50)
	t.Cleanup(func() {
		clearExchangeCodes(t)
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/exchange-codes", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK   bool `json:"ok"`
		Data struct {
			Codes []struct {
				ID       uint32 `json:"id"`
				Code     string `json:"code"`
				Platform string `json:"platform"`
				Quota    int    `json:"quota"`
				Rewards  []struct {
					Type  uint32 `json:"type"`
					ID    uint32 `json:"id"`
					Count uint32 `json:"count"`
				} `json:"rewards"`
			} `json:"codes"`
			Meta struct {
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
	if len(responseStruct.Data.Codes) != 2 {
		t.Fatalf("expected 2 codes, got %d", len(responseStruct.Data.Codes))
	}
	if responseStruct.Data.Meta.Total != 2 {
		t.Fatalf("expected total 2, got %d", responseStruct.Data.Meta.Total)
	}
	if responseStruct.Data.Codes[0].Code != "TESTCODE1" {
		t.Fatalf("expected first code 'TESTCODE1', got %s", responseStruct.Data.Codes[0].Code)
	}
}

func TestExchangeCodeDetail(t *testing.T) {
	app := newExchangeCodeTestApp(t)
	clearExchangeCodes(t)
	seedExchangeCode(t, 1, "TESTCODE3", "ios", 200)
	t.Cleanup(func() {
		clearExchangeCodes(t)
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/exchange-codes/1", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK   bool `json:"ok"`
		Data struct {
			ID       uint32 `json:"id"`
			Code     string `json:"code"`
			Platform string `json:"platform"`
			Quota    int    `json:"quota"`
			Rewards  []struct {
				Type  uint32 `json:"type"`
				ID    uint32 `json:"id"`
				Count uint32 `json:"count"`
			} `json:"rewards"`
		} `json:"data"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}
	if responseStruct.Data.ID != 1 {
		t.Fatalf("expected id 1, got %d", responseStruct.Data.ID)
	}
	if responseStruct.Data.Code != "TESTCODE3" {
		t.Fatalf("expected code 'TESTCODE3', got %s", responseStruct.Data.Code)
	}
	if responseStruct.Data.Platform != "ios" {
		t.Fatalf("expected platform 'ios', got %s", responseStruct.Data.Platform)
	}
	if responseStruct.Data.Quota != 200 {
		t.Fatalf("expected quota 200, got %d", responseStruct.Data.Quota)
	}
}

func TestExchangeCodeDetailNotFound(t *testing.T) {
	app := newExchangeCodeTestApp(t)
	clearExchangeCodes(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/exchange-codes/999", nil)
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

func TestCreateExchangeCode(t *testing.T) {
	app := newExchangeCodeTestApp(t)
	clearExchangeCodes(t)
	t.Cleanup(func() {
		clearExchangeCodes(t)
	})

	requestBody := `{"code":"NEWCODE","platform":"all","quota":100,"rewards":[{"type":1,"id":1001,"count":100}]}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/exchange-codes", strings.NewReader(requestBody))
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

	var code orm.ExchangeCode
	if err := orm.GormDB.Where("code = ?", "NEWCODE").First(&code).Error; err != nil {
		t.Fatalf("query code failed: %v", err)
	}
	if code.Code != "NEWCODE" {
		t.Fatalf("expected code 'NEWCODE', got %s", code.Code)
	}
	if code.Platform != "all" {
		t.Fatalf("expected platform 'all', got %s", code.Platform)
	}
	if code.Quota != 100 {
		t.Fatalf("expected quota 100, got %d", code.Quota)
	}
}

func TestCreateExchangeCodeMissingCode(t *testing.T) {
	app := newExchangeCodeTestApp(t)
	clearExchangeCodes(t)

	requestBody := `{"platform":"all","quota":100,"rewards":[]}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/exchange-codes", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}

	var responseStruct struct {
		OK bool `json:"ok"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if responseStruct.OK {
		t.Fatalf("expected ok false for missing code")
	}
}

func TestUpdateExchangeCode(t *testing.T) {
	app := newExchangeCodeTestApp(t)
	clearExchangeCodes(t)
	seedExchangeCode(t, 1, "TESTCODE4", "android", 100)
	t.Cleanup(func() {
		clearExchangeCodes(t)
	})

	requestBody := `{"code":"UPDATEDCODE","platform":"ios","quota":200,"rewards":[{"type":2,"id":2001,"count":200}]}`
	request := httptest.NewRequest(http.MethodPut, "/api/v1/exchange-codes/1", strings.NewReader(requestBody))
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

	var code orm.ExchangeCode
	if err := orm.GormDB.First(&code, 1).Error; err != nil {
		t.Fatalf("query code failed: %v", err)
	}
	if code.Code != "UPDATEDCODE" {
		t.Fatalf("expected code 'UPDATEDCODE', got %s", code.Code)
	}
	if code.Platform != "ios" {
		t.Fatalf("expected platform 'ios', got %s", code.Platform)
	}
	if code.Quota != 200 {
		t.Fatalf("expected quota 200 after update, got %d", code.Quota)
	}
}

func TestUpdateExchangeCodeNotFound(t *testing.T) {
	app := newExchangeCodeTestApp(t)
	clearExchangeCodes(t)

	requestBody := `{"code":"UPDATEDCODE","platform":"ios","quota":200,"rewards":[]}`
	request := httptest.NewRequest(http.MethodPut, "/api/v1/exchange-codes/999", strings.NewReader(requestBody))
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

func TestDeleteExchangeCode(t *testing.T) {
	app := newExchangeCodeTestApp(t)
	clearExchangeCodes(t)
	seedExchangeCode(t, 1, "TESTCODE5", "all", 100)
	t.Cleanup(func() {
		clearExchangeCodes(t)
	})

	request := httptest.NewRequest(http.MethodDelete, "/api/v1/exchange-codes/1", nil)
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

	var code orm.ExchangeCode
	if err := orm.GormDB.First(&code, 1).Error; err == nil {
		t.Fatalf("expected code to be deleted")
	}
}

func TestDeleteExchangeCodeNotFound(t *testing.T) {
	app := newExchangeCodeTestApp(t)
	clearExchangeCodes(t)

	request := httptest.NewRequest(http.MethodDelete, "/api/v1/exchange-codes/999", nil)
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
