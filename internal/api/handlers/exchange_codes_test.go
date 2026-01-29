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

func clearExchangeCodeRedeems(t *testing.T) {
	t.Helper()
	if err := orm.GormDB.Exec("DELETE FROM exchange_code_redeems").Error; err != nil {
		t.Fatalf("clear exchange code redeems: %v", err)
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

func TestExchangeCodeRedeemFlow(t *testing.T) {
	app := newExchangeCodeTestApp(t)
	clearExchangeCodeRedeems(t)
	clearExchangeCodes(t)
	clearCommanders(t)
	seedExchangeCode(t, 10, "REDEEM10", "all", 1)
	seedCommander(t, 9001, "Redeem Commander")
	defer func() {
		clearExchangeCodeRedeems(t)
		clearExchangeCodes(t)
		clearCommanders(t)
	}()

	createBody := `{"commander_id":9001}`
	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/exchange-codes/10/redeems", strings.NewReader(createBody))
	createRequest.Header.Set("Content-Type", "application/json")
	createResponse := httptest.NewRecorder()
	app.ServeHTTP(createResponse, createRequest)
	if createResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", createResponse.Code)
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/api/v1/exchange-codes/10/redeems", nil)
	listResponse := httptest.NewRecorder()
	app.ServeHTTP(listResponse, listRequest)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", listResponse.Code)
	}
	var listPayload struct {
		OK   bool `json:"ok"`
		Data struct {
			Redeems []struct {
				CommanderID uint32 `json:"commander_id"`
			} `json:"redeems"`
		} `json:"data"`
	}
	if err := json.NewDecoder(listResponse.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listPayload.Data.Redeems) != 1 {
		t.Fatalf("expected 1 redeem, got %d", len(listPayload.Data.Redeems))
	}

	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/v1/exchange-codes/10/redeems/9001", nil)
	deleteResponse := httptest.NewRecorder()
	app.ServeHTTP(deleteResponse, deleteRequest)
	if deleteResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", deleteResponse.Code)
	}
}

func TestExchangeCodeRedeemErrors(t *testing.T) {
	app := newExchangeCodeTestApp(t)
	clearExchangeCodeRedeems(t)
	clearExchangeCodes(t)
	clearCommanders(t)
	seedExchangeCode(t, 11, "REDEEM11", "all", 1)
	seedCommander(t, 9002, "Redeem Commander")
	defer func() {
		clearExchangeCodeRedeems(t)
		clearExchangeCodes(t)
		clearCommanders(t)
	}()

	request := httptest.NewRequest(http.MethodPost, "/api/v1/exchange-codes/11/redeems", strings.NewReader(`{"commander_id":0}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodPost, "/api/v1/exchange-codes/11/redeems", strings.NewReader(`{"commander_id":9002}`))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	duplicateRequest := httptest.NewRequest(http.MethodPost, "/api/v1/exchange-codes/11/redeems", strings.NewReader(`{"commander_id":9002}`))
	duplicateRequest.Header.Set("Content-Type", "application/json")
	duplicateResponse := httptest.NewRecorder()
	app.ServeHTTP(duplicateResponse, duplicateRequest)
	if duplicateResponse.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", duplicateResponse.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/exchange-codes/9999/redeems", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/exchange-codes/0/redeems", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodDelete, "/api/v1/exchange-codes/11/redeems/9999", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodDelete, "/api/v1/exchange-codes/0/redeems/9002", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}
