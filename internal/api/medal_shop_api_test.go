package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
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

const (
	monthShopConfigCategory = "ShareCfg/month_shop_template.json"
	shopTemplateCategory    = "ShareCfg/shop_template.json"
)

type medalMonthShopTemplate struct {
	HonorMedalShopGoods []uint32 `json:"honormedal_shop_goods"`
}

type medalShopTemplateEntry struct {
	ID                 uint32 `json:"id"`
	GoodsPurchaseLimit uint32 `json:"goods_purchase_limit"`
}

func seedMedalShopConfig(t *testing.T) {
	orm.GormDB.Where("category = ?", monthShopConfigCategory).Delete(&orm.ConfigEntry{})
	orm.GormDB.Where("category = ?", shopTemplateCategory).Delete(&orm.ConfigEntry{})
	monthPayload, err := json.Marshal(medalMonthShopTemplate{HonorMedalShopGoods: []uint32{10000, 10001}})
	if err != nil {
		t.Fatalf("failed to marshal month shop template: %v", err)
	}
	if err := orm.GormDB.Create(&orm.ConfigEntry{Category: monthShopConfigCategory, Key: "1", Data: monthPayload}).Error; err != nil {
		t.Fatalf("failed to create month shop entry: %v", err)
	}
	entries := []medalShopTemplateEntry{{ID: 10000, GoodsPurchaseLimit: 5}, {ID: 10001, GoodsPurchaseLimit: 2}}
	for _, entry := range entries {
		payload, err := json.Marshal(entry)
		if err != nil {
			t.Fatalf("failed to marshal shop template entry: %v", err)
		}
		if err := orm.GormDB.Create(&orm.ConfigEntry{Category: shopTemplateCategory, Key: fmt.Sprintf("%d", entry.ID), Data: payload}).Error; err != nil {
			t.Fatalf("failed to create shop template entry: %v", err)
		}
	}
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
