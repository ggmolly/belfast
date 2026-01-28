package api_test

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

type shoppingStreetResponse struct {
	OK    bool                         `json:"ok"`
	Data  types.ShoppingStreetResponse `json:"data"`
	Error *types.APIError              `json:"error,omitempty"`
}

type shoppingStreetOfferListResponse struct {
	OK   bool                        `json:"ok"`
	Data types.ShopOfferListResponse `json:"data"`
}

func resetShoppingStreetData(t *testing.T) {
	if err := orm.GormDB.Exec("DELETE FROM shopping_street_goods").Error; err != nil {
		t.Fatalf("failed to clear shopping_street_goods: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM shopping_street_states").Error; err != nil {
		t.Fatalf("failed to clear shopping_street_states: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM shop_offers").Error; err != nil {
		t.Fatalf("failed to clear shop_offers: %v", err)
	}
}

func apiSeedShoppingStreetOffers(t *testing.T) []orm.ShopOffer {
	offers := []orm.ShopOffer{
		{
			ID:             910001,
			Effects:        orm.Int64List{1},
			EffectArgs:     json.RawMessage("[1]"),
			Number:         1,
			ResourceNumber: 10,
			ResourceID:     1,
			Type:           1,
			Genre:          "shopping_street",
			Discount:       10,
		},
		{
			ID:             910002,
			Effects:        orm.Int64List{2},
			EffectArgs:     json.RawMessage("[2]"),
			Number:         1,
			ResourceNumber: 15,
			ResourceID:     1,
			Type:           1,
			Genre:          "shopping_street",
			Discount:       0,
		},
		{
			ID:             910003,
			Effects:        orm.Int64List{3},
			EffectArgs:     json.RawMessage("[3]"),
			Number:         1,
			ResourceNumber: 20,
			ResourceID:     1,
			Type:           1,
			Genre:          "regular",
			Discount:       0,
		},
	}
	for _, offer := range offers {
		if err := orm.GormDB.Create(&offer).Error; err != nil {
			t.Fatalf("failed to create shop offer: %v", err)
		}
	}
	return offers
}

func TestShoppingStreetGetAndFilterOffers(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)
	resetShoppingStreetData(t)
	offers := apiSeedShoppingStreetOffers(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players/1/shopping-street?include_offers=true", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload shoppingStreetResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !payload.OK {
		t.Fatalf("expected ok response")
	}
	if len(payload.Data.Goods) != 2 {
		t.Fatalf("expected 2 goods, got %d", len(payload.Data.Goods))
	}
	if payload.Data.Goods[0].Offer == nil {
		t.Fatalf("expected offer metadata")
	}
	if payload.Data.State.NextFlashTime <= uint32(time.Now().Unix()) {
		t.Fatalf("expected next flash time in the future")
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/shop/offers?genre=shopping_street&offset=0&limit=10", nil)
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	var offerPayload shoppingStreetOfferListResponse
	if err := json.NewDecoder(response.Body).Decode(&offerPayload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if offerPayload.Data.Meta.Total != 2 {
		t.Fatalf("expected 2 offers, got %d", offerPayload.Data.Meta.Total)
	}
	_ = offers
}

func TestShoppingStreetRefreshAndStateUpdates(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)
	resetShoppingStreetData(t)
	offers := apiSeedShoppingStreetOffers(t)

	payload := []byte(`{"goods_ids":[999999],"goods_count":1}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/players/1/shopping-street/refresh", bytes.NewBuffer(payload))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}

	payload = []byte(`{"goods_ids":[910001],"discount_override":90,"buy_count":2,"next_flash_in_seconds":10}`)
	request = httptest.NewRequest(http.MethodPost, "/api/v1/players/1/shopping-street/refresh", bytes.NewBuffer(payload))
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	var refreshPayload shoppingStreetResponse
	if err := json.NewDecoder(response.Body).Decode(&refreshPayload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(refreshPayload.Data.Goods) != 1 {
		t.Fatalf("expected 1 good, got %d", len(refreshPayload.Data.Goods))
	}
	if refreshPayload.Data.Goods[0].Discount != 90 {
		t.Fatalf("expected discount 90, got %d", refreshPayload.Data.Goods[0].Discount)
	}
	if refreshPayload.Data.Goods[0].BuyCount != 2 {
		t.Fatalf("expected buy_count 2, got %d", refreshPayload.Data.Goods[0].BuyCount)
	}
	if refreshPayload.Data.State.NextFlashTime <= uint32(time.Now().Unix()) {
		t.Fatalf("expected next flash time in the future")
	}

	payload = []byte(`{"level":3,"flash_count":5}`)
	request = httptest.NewRequest(http.MethodPut, "/api/v1/players/1/shopping-street", bytes.NewBuffer(payload))
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	var updatePayload shoppingStreetResponse
	if err := json.NewDecoder(response.Body).Decode(&updatePayload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if updatePayload.Data.State.Level != 3 {
		t.Fatalf("expected level 3, got %d", updatePayload.Data.State.Level)
	}
	if updatePayload.Data.State.FlashCount != 5 {
		t.Fatalf("expected flash_count 5, got %d", updatePayload.Data.State.FlashCount)
	}
	_ = offers
}

func TestShoppingStreetGoodsMutation(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)
	resetShoppingStreetData(t)
	apiSeedShoppingStreetOffers(t)

	payload := []byte(`{"goods":[{"goods_id":99999,"discount":80,"buy_count":1}]}`)
	request := httptest.NewRequest(http.MethodPut, "/api/v1/players/1/shopping-street/goods", bytes.NewBuffer(payload))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}

	payload = []byte(`{"goods":[{"goods_id":910002,"discount":85,"buy_count":3}]}`)
	request = httptest.NewRequest(http.MethodPut, "/api/v1/players/1/shopping-street/goods", bytes.NewBuffer(payload))
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	var replacePayload shoppingStreetResponse
	if err := json.NewDecoder(response.Body).Decode(&replacePayload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(replacePayload.Data.Goods) != 1 {
		t.Fatalf("expected 1 good, got %d", len(replacePayload.Data.Goods))
	}
	if replacePayload.Data.Goods[0].BuyCount != 3 {
		t.Fatalf("expected buy_count 3, got %d", replacePayload.Data.Goods[0].BuyCount)
	}

	payload = []byte(`{"buy_count":2}`)
	request = httptest.NewRequest(http.MethodPatch, "/api/v1/players/1/shopping-street/goods/910002", bytes.NewBuffer(payload))
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	var patchPayload shoppingStreetResponse
	if err := json.NewDecoder(response.Body).Decode(&patchPayload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if patchPayload.Data.Goods[0].BuyCount != 2 {
		t.Fatalf("expected buy_count 2, got %d", patchPayload.Data.Goods[0].BuyCount)
	}

	request = httptest.NewRequest(http.MethodDelete, "/api/v1/players/1/shopping-street/goods/910002", nil)
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	var deletePayload shoppingStreetResponse
	if err := json.NewDecoder(response.Body).Decode(&deletePayload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(deletePayload.Data.Goods) != 0 {
		t.Fatalf("expected 0 goods, got %d", len(deletePayload.Data.Goods))
	}
}
