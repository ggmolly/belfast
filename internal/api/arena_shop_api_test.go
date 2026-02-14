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

type arenaShopResponse struct {
	OK    bool                    `json:"ok"`
	Data  types.ArenaShopResponse `json:"data"`
	Error *types.APIError         `json:"error,omitempty"`
}

const arenaShopConfigCategory = "ShareCfg/arena_data_shop.json"

type arenaShopTemplate struct {
	CommodityList1      [][]uint32 `json:"commodity_list_1"`
	CommodityList2      [][]uint32 `json:"commodity_list_2"`
	CommodityList3      [][]uint32 `json:"commodity_list_3"`
	CommodityList4      [][]uint32 `json:"commodity_list_4"`
	CommodityList5      [][]uint32 `json:"commodity_list_5"`
	CommodityListCommon [][]uint32 `json:"commodity_list_common"`
	RefreshPrice        []uint32   `json:"refresh_price"`
}

func resetArenaShopData(t *testing.T) {
	execAPITestSQLT(t, "DELETE FROM arena_shop_states")
	execAPITestSQLT(t, "DELETE FROM config_entries WHERE category = $1", arenaShopConfigCategory)
}

func seedArenaShopConfigEntry(t *testing.T) {
	data, err := json.Marshal(arenaShopTemplate{
		CommodityList1:      [][]uint32{{1001, 1}},
		CommodityList2:      [][]uint32{{1002, 1}},
		CommodityList3:      [][]uint32{{1003, 2}},
		CommodityList4:      [][]uint32{{1004, 3}},
		CommodityList5:      [][]uint32{{1005, 4}},
		CommodityListCommon: [][]uint32{{2001, 5}, {2002, 6}},
		RefreshPrice:        []uint32{20, 50, 100},
	})
	if err != nil {
		t.Fatalf("failed to marshal shop config: %v", err)
	}
	entry := orm.ConfigEntry{
		Category: arenaShopConfigCategory,
		Key:      "1",
		Data:     data,
	}
	if err := orm.CreateConfigEntryRecord(&entry); err != nil {
		t.Fatalf("failed to create config entry: %v", err)
	}
}

func seedArenaShopGems(t *testing.T, amount uint32) {
	execAPITestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3) ON CONFLICT (commander_id, resource_id) DO UPDATE SET amount = EXCLUDED.amount", int64(1), int64(4), int64(amount))
}

func TestArenaShopAPIGet(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)
	resetArenaShopData(t)
	seedArenaShopConfigEntry(t)
	seedArenaShopGems(t, 200)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players/1/arena-shop", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	var payload arenaShopResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !payload.OK {
		t.Fatalf("expected ok response")
	}
	if payload.Data.State.FlashCount != 0 {
		t.Fatalf("expected flash_count 0, got %d", payload.Data.State.FlashCount)
	}
	if len(payload.Data.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(payload.Data.Items))
	}
	if payload.Data.State.NextFlashTime <= uint32(time.Now().Unix()) {
		t.Fatalf("expected next flash time in the future")
	}
}

func TestArenaShopAPIRefresh(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)
	resetArenaShopData(t)
	seedArenaShopConfigEntry(t)
	seedArenaShopGems(t, 100)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/players/1/arena-shop/refresh", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	var payload arenaShopResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if payload.Data.State.FlashCount != 1 {
		t.Fatalf("expected flash_count 1, got %d", payload.Data.State.FlashCount)
	}
	if len(payload.Data.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(payload.Data.Items))
	}
	amount := queryAPITestInt64(t, "SELECT amount FROM owned_resources WHERE commander_id = $1 AND resource_id = $2", int64(1), int64(4))
	if amount != 80 {
		t.Fatalf("expected gem amount 80, got %d", amount)
	}
}

func TestArenaShopAPIUpdateState(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)
	resetArenaShopData(t)
	seedArenaShopConfigEntry(t)
	seedArenaShopGems(t, 50)

	payload := []byte(`{"flash_count":2,"next_flash_time":12345,"last_refresh_time":900}`)
	request := httptest.NewRequest(http.MethodPut, "/api/v1/players/1/arena-shop", bytes.NewBuffer(payload))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	var responsePayload arenaShopResponse
	if err := json.NewDecoder(response.Body).Decode(&responsePayload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if responsePayload.Data.State.FlashCount != 2 {
		t.Fatalf("expected flash_count 2, got %d", responsePayload.Data.State.FlashCount)
	}
	if responsePayload.Data.State.NextFlashTime != 12345 {
		t.Fatalf("expected next_flash_time 12345, got %d", responsePayload.Data.State.NextFlashTime)
	}
	if responsePayload.Data.State.LastRefreshTime != 900 {
		t.Fatalf("expected last_refresh_time 900, got %d", responsePayload.Data.State.LastRefreshTime)
	}
	if len(responsePayload.Data.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(responsePayload.Data.Items))
	}
}
