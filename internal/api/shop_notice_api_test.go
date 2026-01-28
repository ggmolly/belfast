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

type shopOfferListResponse struct {
	OK   bool                        `json:"ok"`
	Data types.ShopOfferListResponse `json:"data"`
}

type noticeListResponse struct {
	OK   bool                     `json:"ok"`
	Data types.NoticeListResponse `json:"data"`
}

func TestShopOfferList(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	offer := orm.ShopOffer{
		ID:             1,
		Effects:        orm.Int64List{100},
		Number:         2,
		ResourceNumber: 10,
		ResourceID:     1,
		Type:           1,
	}
	if err := orm.GormDB.Create(&offer).Error; err != nil {
		t.Fatalf("failed to create offer: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/shop/offers?offset=0&limit=10", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload shopOfferListResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !payload.OK {
		t.Fatalf("expected ok response")
	}
	if payload.Data.Meta.Total != 1 {
		t.Fatalf("expected 1 offer, got %d", payload.Data.Meta.Total)
	}
}

func TestShopOfferCreateUpdateDelete(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	payload := []byte(`{"id":2,"effects":[200],"num":5,"resource_num":15,"resource_type":1,"type":1}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/shop/offers", bytes.NewBuffer(payload))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	payload = []byte(`{"effects":[201],"num":6,"resource_num":20,"resource_type":2,"type":2}`)
	request = httptest.NewRequest(http.MethodPut, "/api/v1/shop/offers/2", bytes.NewBuffer(payload))
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodDelete, "/api/v1/shop/offers/2", nil)
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
}

func TestNoticeListAndActive(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	notice := orm.Notice{ID: 1, Version: "1", BtnTitle: "OK", Title: "Title", TitleImage: "img", TimeDesc: "Now", Content: "Body", TagType: 1, Icon: 1, Track: "main"}
	if err := orm.GormDB.Create(&notice).Error; err != nil {
		t.Fatalf("failed to create notice: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/notices?offset=0&limit=10", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload noticeListResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if payload.Data.Meta.Total != 1 {
		t.Fatalf("expected 1 notice, got %d", payload.Data.Meta.Total)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/notices/active", nil)
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
}

func TestNoticeCreateUpdateDelete(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	payload := []byte(`{"id":2,"version":"1","btn_title":"OK","title":"Hello","title_image":"img","time_desc":"Now","content":"Body","tag_type":1,"icon":1,"track":"main"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/notices", bytes.NewBuffer(payload))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	payload = []byte(`{"version":"2","btn_title":"OK","title":"Updated","title_image":"img","time_desc":"Later","content":"Body","tag_type":2,"icon":2,"track":"main"}`)
	request = httptest.NewRequest(http.MethodPut, "/api/v1/notices/2", bytes.NewBuffer(payload))
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodDelete, "/api/v1/notices/2", nil)
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
}
