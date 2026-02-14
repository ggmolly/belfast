package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api"
	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
)

type playerListResponse struct {
	OK   bool                     `json:"ok"`
	Data types.PlayerListResponse `json:"data"`
}

type genericResponse struct {
	OK bool `json:"ok"`
}

type playerCompensationResponse struct {
	OK   bool                             `json:"ok"`
	Data types.PlayerCompensationResponse `json:"data"`
}

type playerBuildQueueResponse struct {
	OK   bool                           `json:"ok"`
	Data types.PlayerBuildQueueResponse `json:"data"`
}

type playerSkinResponse struct {
	OK   bool                     `json:"ok"`
	Data types.PlayerSkinResponse `json:"data"`
}

func TestPlayerListFilters(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players?offset=0&limit=10", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload playerListResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !payload.OK {
		t.Fatalf("expected ok response")
	}
	if payload.Data.Meta.Total < 2 {
		t.Fatalf("expected at least 2 players, got %d", payload.Data.Meta.Total)
	}
}

func TestCreatePlayerInvalidTimestampsRollback(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	before := queryAPITestInt64(t, "SELECT COUNT(*) FROM commanders")

	body := []byte(`{"commander_id":3,"account_id":12,"name":"BadTimestamps","last_login":"not-a-time","name_change_cooldown":"not-a-time"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/players", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}

	count := queryAPITestInt64(t, "SELECT COUNT(*) FROM commanders WHERE commander_id = $1", int64(3))
	if count != 0 {
		t.Fatalf("expected no commander persisted for invalid timestamps, got %d", count)
	}

	after := queryAPITestInt64(t, "SELECT COUNT(*) FROM commanders")
	if after != before {
		t.Fatalf("expected no net command table change, before %d after %d", before, after)
	}
}

func TestPlayerBanUnban(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	body := []byte(`{"duration_sec": 60}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/players/1/ban", bytes.NewBuffer(body))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var count int64
	count = queryAPITestInt64(t, "SELECT COUNT(*) FROM punishments WHERE punished_id = $1", int64(1))
	if count == 0 {
		t.Fatalf("expected punishment to be created")
	}

	request = httptest.NewRequest(http.MethodDelete, "/api/v1/players/1/ban", nil)
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodDelete, "/api/v1/players/1/ban", nil)
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", response.Code)
	}
}

func TestPlayerResourcesSet(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	body := []byte(`{"resources":[{"resource_id":1,"amount":500}]}`)
	request := httptest.NewRequest(http.MethodPut, "/api/v1/players/1/resources", bytes.NewBuffer(body))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	amount := queryAPITestInt64(t, "SELECT amount FROM owned_resources WHERE commander_id = $1 AND resource_id = $2", int64(1), int64(1))
	if amount != 500 {
		t.Fatalf("expected amount 500, got %d", amount)
	}
}

func TestPlayerGiveItemShipMail(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	itemBody := []byte(`{"item_id":20001,"amount":2}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/players/1/give-item", bytes.NewBuffer(itemBody))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	shipBody := []byte(`{"ship_id":1}`)
	request = httptest.NewRequest(http.MethodPost, "/api/v1/players/1/give-ship", bytes.NewBuffer(shipBody))
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	mailBody := []byte(`{"title":"Hello","body":"Test","attachments":[{"type":2,"item_id":20001,"quantity":1}]}`)
	request = httptest.NewRequest(http.MethodPost, "/api/v1/players/1/send-mail", bytes.NewBuffer(mailBody))
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
}

func TestPlayerCompensationCrud(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	sendTime := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	expiresAt := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)
	createBody := fmt.Sprintf(`{"title":"Apology","text":"Test","send_time":"%s","expires_at":"%s","attachments":[{"type":2,"item_id":20001,"quantity":1}]}`, sendTime, expiresAt)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/players/1/compensations", bytes.NewBuffer([]byte(createBody)))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	compensationID := queryAPITestInt64(t, "SELECT id FROM compensations WHERE commander_id = $1 ORDER BY id LIMIT 1", int64(1))

	request = httptest.NewRequest(http.MethodGet, "/api/v1/players/1/compensations", nil)
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	var listPayload playerCompensationResponse
	if err := json.NewDecoder(response.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !listPayload.OK {
		t.Fatalf("expected ok response")
	}
	if len(listPayload.Data.Compensations) != 1 {
		t.Fatalf("expected 1 compensation, got %d", len(listPayload.Data.Compensations))
	}
	entry := listPayload.Data.Compensations[0]
	if entry.CompensationID != uint32(compensationID) {
		t.Fatalf("expected compensation_id %d, got %d", compensationID, entry.CompensationID)
	}
	if entry.Title != "Apology" {
		t.Fatalf("expected title Apology, got %s", entry.Title)
	}
	if entry.Text != "Test" {
		t.Fatalf("expected text Test, got %s", entry.Text)
	}
	if entry.SendTime == "" || entry.ExpiresAt == "" {
		t.Fatalf("expected send_time and expires_at")
	}
	if len(entry.Attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(entry.Attachments))
	}
	if entry.Attachments[0].Type != 2 || entry.Attachments[0].ItemID != 20001 || entry.Attachments[0].Quantity != 1 {
		t.Fatalf("unexpected attachment payload: %+v", entry.Attachments[0])
	}

	request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/players/1/compensations/%d", compensationID), nil)
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	updateBody := `{"title":"Updated","attach_flag":true,"attachments":[{"type":1,"item_id":1,"quantity":5}]}`
	request = httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/players/1/compensations/%d", compensationID), bytes.NewBuffer([]byte(updateBody)))
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var compensationTitle string
	var attachFlag bool
	if err := db.DefaultStore.Pool.QueryRow(context.Background(), "SELECT title, attach_flag FROM compensations WHERE id = $1", compensationID).Scan(&compensationTitle, &attachFlag); err != nil {
		t.Fatalf("failed to reload compensation: %v", err)
	}
	if compensationTitle != "Updated" {
		t.Fatalf("expected title updated, got %s", compensationTitle)
	}
	if !attachFlag {
		t.Fatalf("expected attach flag true")
	}
	attachmentCount := queryAPITestInt64(t, "SELECT COUNT(*) FROM compensation_attachments WHERE compensation_id = $1", compensationID)
	if attachmentCount != 1 {
		t.Fatalf("expected 1 attachment, got %d", attachmentCount)
	}

	request = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/players/1/compensations/%d", compensationID), nil)
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	count := queryAPITestInt64(t, "SELECT COUNT(*) FROM compensations WHERE commander_id = $1", int64(1))
	if count != 0 {
		t.Fatalf("expected 0 compensations, got %d", count)
	}
}

func TestPlayerCompensationPush(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	expiresAt := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)
	createBody := fmt.Sprintf(`{"title":"Notice","text":"Push","expires_at":"%s"}`, expiresAt)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/players/1/compensations", bytes.NewBuffer([]byte(createBody)))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	client := &connection.Client{Commander: &orm.Commander{CommanderID: 1}, Server: connection.BelfastInstance, Hash: 9001}
	connection.BelfastInstance.AddClient(client)
	defer connection.BelfastInstance.RemoveClient(client)

	request = httptest.NewRequest(http.MethodPost, "/api/v1/players/1/compensations/push", nil)
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	notification := &protobuf.SC_30101{}
	decodeTestPacket(t, client.Buffer.Bytes(), 30101, notification)
	if notification.GetNumber() != 1 {
		t.Fatalf("expected number 1, got %d", notification.GetNumber())
	}
}

func TestPlayerCompensationPushOnline(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	expiresAt := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)
	createBody := fmt.Sprintf(`{"title":"Notice","text":"Push","expires_at":"%s"}`, expiresAt)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/players/1/compensations", bytes.NewBuffer([]byte(createBody)))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	client := &connection.Client{Commander: &orm.Commander{CommanderID: 1}, Server: connection.BelfastInstance, Hash: 9002}
	connection.BelfastInstance.AddClient(client)
	defer connection.BelfastInstance.RemoveClient(client)

	request = httptest.NewRequest(http.MethodPost, "/api/v1/players/compensations/push-online", nil)
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload struct {
		OK   bool                           `json:"ok"`
		Data types.PushCompensationResponse `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if payload.Data.Pushed != 1 {
		t.Fatalf("expected pushed 1, got %d", payload.Data.Pushed)
	}

	notification := &protobuf.SC_30101{}
	decodeTestPacket(t, client.Buffer.Bytes(), 30101, notification)
	if notification.GetNumber() != 1 {
		t.Fatalf("expected number 1, got %d", notification.GetNumber())
	}
}

func TestPlayerDeleteSoft(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	request := httptest.NewRequest(http.MethodDelete, "/api/v1/players/1", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	deleted := queryAPITestInt64(t, "SELECT COUNT(*) FROM commanders WHERE commander_id = $1 AND deleted_at IS NOT NULL", int64(1))
	if deleted == 0 {
		t.Fatalf("expected deleted_at to be set")
	}
}

var testApp *iris.Application

func setupTestAPI(t *testing.T) {
	os.Setenv("MODE", "test")
	if ok := orm.InitDatabase(); !ok {
	}
	cfg := api.Config{Enabled: true, Port: 0}
	cfg.RuntimeConfig = &config.Config{}
	cfg.RuntimeConfig.Auth.DisableAuth = true
	connection.BelfastInstance = connection.NewServer("127.0.0.1", 0, func(*[]byte, *connection.Client, int) {})
	if testApp != nil {
		return
	}
	app := api.NewApp(cfg)
	app.Build()
	testApp = app
}

func seedPlayers(t *testing.T) {
	execAPITestSQLT(t, "DELETE FROM punishments")
	execAPITestSQLT(t, "DELETE FROM owned_resources")
	execAPITestSQLT(t, "DELETE FROM commander_items")
	execAPITestSQLT(t, "DELETE FROM commander_misc_items")
	execAPITestSQLT(t, "DELETE FROM owned_ships")
	execAPITestSQLT(t, "DELETE FROM builds")
	execAPITestSQLT(t, "DELETE FROM mails")
	execAPITestSQLT(t, "DELETE FROM mail_attachments")
	execAPITestSQLT(t, "DELETE FROM compensations")
	execAPITestSQLT(t, "DELETE FROM compensation_attachments")
	execAPITestSQLT(t, "DELETE FROM fleets")
	execAPITestSQLT(t, "DELETE FROM owned_skins")
	execAPITestSQLT(t, "DELETE FROM commanders")
	execAPITestSQLT(t, "DELETE FROM resources")
	execAPITestSQLT(t, "DELETE FROM items")
	execAPITestSQLT(t, "DELETE FROM ships")
	execAPITestSQLT(t, "DELETE FROM skins")

	seedDb()

	execAPITestSQLT(t, "INSERT INTO items (id, name, rarity, shop_id, type, virtual_type) VALUES ($1, $2, $3, $4, $5, $6)", int64(30001), "Misc Item", int64(1), int64(0), int64(0), int64(0))

	if err := orm.CreateCommanderRoot(1, 10, "Alpha", 0, 0); err != nil {
		t.Fatalf("failed to create commander1: %v", err)
	}
	if err := orm.CreateCommanderRoot(2, 11, "Bravo", 0, 0); err != nil {
		t.Fatalf("failed to create commander2: %v", err)
	}
	execAPITestSQLT(t, "UPDATE commanders SET level = $1, last_login = $2 WHERE commander_id = $3", int64(5), time.Now().Add(-time.Hour), int64(1))
	execAPITestSQLT(t, "UPDATE commanders SET level = $1, last_login = $2 WHERE commander_id = $3", int64(20), time.Now().Add(-2*time.Hour), int64(2))

	execAPITestSQLT(t, "INSERT INTO ships (template_id, name, rarity_id, star, type, english_name, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(1), "Test Ship", int64(2), int64(1), int64(1), "Test Ship", int64(1), int64(1))

	execAPITestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(1), int64(1), int64(100))
	execAPITestSQLT(t, "INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)", int64(1), int64(2), int64(30))

	execAPITestSQLT(t, "INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)", int64(1), int64(20001), int64(5))
	execAPITestSQLT(t, "INSERT INTO commander_misc_items (commander_id, item_id, data) VALUES ($1, $2, $3)", int64(1), int64(30001), int64(2))
}

func TestPlayerSearch(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players/search?q=alp&offset=0&limit=10", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload playerListResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(payload.Data.Players) != 1 {
		t.Fatalf("expected 1 player, got %d", len(payload.Data.Players))
	}
}

func TestPlayerResourcesList(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players/1/resources", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload struct {
		OK   bool                         `json:"ok"`
		Data types.PlayerResourceResponse `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(payload.Data.Resources) == 0 {
		t.Fatalf("expected resources")
	}
}

func TestPlayerItemsList(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players/1/items", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload struct {
		OK   bool                     `json:"ok"`
		Data types.PlayerItemResponse `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(payload.Data.Items) == 0 {
		t.Fatalf("expected items")
	}
}

func TestPlayerDetail(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players/1", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload struct {
		OK   bool                       `json:"ok"`
		Data types.PlayerDetailResponse `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if payload.Data.AccountID != 10 {
		t.Fatalf("expected account_id 10, got %d", payload.Data.AccountID)
	}
}

func TestPlayerSkinsList(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	execAPITestSQLT(t, "INSERT INTO skins (id, name, ship_group) VALUES ($1, $2, $3)", int64(1001), "Test Skin A", int64(1))
	execAPITestSQLT(t, "INSERT INTO skins (id, name, ship_group) VALUES ($1, $2, $3)", int64(1002), "Test Skin B", int64(1))
	expiresAt := time.Now().Add(24 * time.Hour)
	execAPITestSQLT(t, "INSERT INTO owned_skins (commander_id, skin_id, expires_at) VALUES ($1, $2, $3)", int64(1), int64(1002), expiresAt)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players/1/skins", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload struct {
		OK   bool                     `json:"ok"`
		Data types.PlayerSkinResponse `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(payload.Data.Skins) != 1 {
		t.Fatalf("expected 1 skin, got %d", len(payload.Data.Skins))
	}
	if payload.Data.Skins[0].SkinID != 1002 {
		t.Fatalf("expected skin_id 1002, got %d", payload.Data.Skins[0].SkinID)
	}
}

func TestPlayerFleetsList(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players/1/fleets", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
}

func TestPlayerFleetCreateBusyShipReturnsConflict(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	execAPITestSQLT(t, "DELETE FROM event_collections")

	now := time.Now().UTC()
	execAPITestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, level, energy, create_time, change_name_timestamp) VALUES ($1, $2, $3, $4, $5, $6, $7)", int64(101), int64(1), int64(1), int64(1), int64(150), now, now)
	execAPITestSQLT(t, "INSERT INTO event_collections (commander_id, collection_id, start_time, finish_time, ship_ids) VALUES ($1, $2, $3, $4, $5::jsonb)", int64(1), int64(1), int64(1), int64(2), "[101]")

	body := []byte(`{"game_id":1,"name":"Main","ship_ids":[101]}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/players/1/fleets", bytes.NewBuffer(body))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", response.Code)
	}
}

func TestPlayerBuildsList(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	execAPITestSQLT(t, "INSERT INTO builds (builder_id, ship_id, pool_id, finishes_at) VALUES ($1, $2, $3, $4)", int64(1), int64(1), int64(1), time.Now().Add(time.Hour))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players/1/builds", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
}

func TestPlayerBuildQueueSnapshot(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	execAPITestSQLT(t, "INSERT INTO builds (builder_id, ship_id, pool_id, finishes_at) VALUES ($1, $2, $3, $4)", int64(1), int64(1), int64(1), time.Now().Add(2*time.Hour))
	execAPITestSQLT(t, "INSERT INTO builds (builder_id, ship_id, pool_id, finishes_at) VALUES ($1, $2, $3, $4)", int64(1), int64(1), int64(2), time.Now().Add(-time.Hour))

	execAPITestSQLT(t, "UPDATE commanders SET draw_count1 = $1, draw_count10 = $2, exchange_count = $3 WHERE commander_id = $4", int64(4), int64(5), int64(6), int64(1))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players/1/builds/queue", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload playerBuildQueueResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !payload.OK {
		t.Fatalf("expected ok response")
	}
	if payload.Data.WorklistCount != consts.MaxBuildWorkCount {
		t.Fatalf("expected worklist count %d, got %d", consts.MaxBuildWorkCount, payload.Data.WorklistCount)
	}
	if len(payload.Data.WorklistList) != 2 {
		t.Fatalf("expected 2 builds, got %d", len(payload.Data.WorklistList))
	}
	if payload.Data.WorklistList[0].PoolID != 1 {
		t.Fatalf("expected pool id %d, got %d", 1, payload.Data.WorklistList[0].PoolID)
	}
	if payload.Data.WorklistList[1].RemainingSeconds != 0 {
		t.Fatalf("expected finished build to have 0 remaining seconds")
	}
	if payload.Data.DrawCount1 != 4 || payload.Data.DrawCount10 != 5 || payload.Data.ExchangeCount != 6 {
		t.Fatalf("unexpected counters in response")
	}
}

func TestPlayerBuildCountersUpdate(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	body := []byte(`{"draw_count_1":2,"draw_count_10":3,"exchange_count":7}`)
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/players/1/builds/counters", bytes.NewBuffer(body))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	drawCount1 := queryAPITestInt64(t, "SELECT draw_count1 FROM commanders WHERE commander_id = $1", int64(1))
	drawCount10 := queryAPITestInt64(t, "SELECT draw_count10 FROM commanders WHERE commander_id = $1", int64(1))
	exchangeCount := queryAPITestInt64(t, "SELECT exchange_count FROM commanders WHERE commander_id = $1", int64(1))
	if drawCount1 != 2 || drawCount10 != 3 || exchangeCount != 7 {
		t.Fatalf("counters did not update")
	}
}

func TestUpdatePlayerBuild(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	execAPITestSQLT(t, "INSERT INTO ships (template_id, name, rarity_id, star, type, english_name, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(2), "Build Ship", int64(2), int64(1), int64(1), "Build Ship", int64(1), int64(1))

	execAPITestSQLT(t, "INSERT INTO builds (id, builder_id, ship_id, pool_id, finishes_at) VALUES ($1, $2, $3, $4, $5)", int64(99), int64(1), int64(1), int64(1), time.Now().Add(2*time.Hour))

	execAPITestSQLT(t, "INSERT INTO builds (id, builder_id, ship_id, pool_id, finishes_at) VALUES ($1, $2, $3, $4, $5)", int64(100), int64(2), int64(1), int64(1), time.Now().Add(2*time.Hour))

	body := []byte(`{"ship_id":2}`)
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/players/1/builds/99", bytes.NewBuffer(body))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	updated, err := orm.GetBuildByID(99)
	if err != nil {
		t.Fatalf("failed to reload build: %v", err)
	}
	if updated.ShipID != 2 {
		t.Fatalf("expected ship_id 2, got %d", updated.ShipID)
	}

	newFinish := time.Now().UTC().Add(3 * time.Hour).Truncate(time.Second)
	body = []byte(fmt.Sprintf(`{"finishes_at":"%s"}`, newFinish.Format(time.RFC3339)))
	request = httptest.NewRequest(http.MethodPatch, "/api/v1/players/1/builds/99", bytes.NewBuffer(body))
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	updated, err = orm.GetBuildByID(99)
	if err != nil {
		t.Fatalf("failed to reload build: %v", err)
	}
	if updated.FinishesAt.UTC().Truncate(time.Second) != newFinish {
		t.Fatalf("expected finishes_at %v, got %v", newFinish, updated.FinishesAt)
	}

	body = []byte(`{"ship_id":99999}`)
	request = httptest.NewRequest(http.MethodPatch, "/api/v1/players/1/builds/99", bytes.NewBuffer(body))
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}

	body = []byte(`{"finishes_at":"bad-date"}`)
	request = httptest.NewRequest(http.MethodPatch, "/api/v1/players/1/builds/99", bytes.NewBuffer(body))
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}

	body = []byte(`{}`)
	request = httptest.NewRequest(http.MethodPatch, "/api/v1/players/1/builds/99", bytes.NewBuffer(body))
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}

	body = []byte(`{"ship_id":2}`)
	request = httptest.NewRequest(http.MethodPatch, "/api/v1/players/1/builds/9999", bytes.NewBuffer(body))
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", response.Code)
	}

	body = []byte(`{"ship_id":2}`)
	request = httptest.NewRequest(http.MethodPatch, "/api/v1/players/1/builds/100", bytes.NewBuffer(body))
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", response.Code)
	}
}

func TestQuickFinishBuild(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	execAPITestSQLT(t, "INSERT INTO builds (id, builder_id, ship_id, pool_id, finishes_at) VALUES ($1, $2, $3, $4, $5)", int64(101), int64(1), int64(1), int64(1), time.Now().Add(2*time.Hour))

	execAPITestSQLT(t, "INSERT INTO builds (id, builder_id, ship_id, pool_id, finishes_at) VALUES ($1, $2, $3, $4, $5)", int64(102), int64(2), int64(1), int64(1), time.Now().Add(2*time.Hour))

	request := httptest.NewRequest(http.MethodPatch, "/api/v1/players/1/builds/101/quick-finish", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	updated, err := orm.GetBuildByID(101)
	if err != nil {
		t.Fatalf("failed to reload build: %v", err)
	}

	expected := time.Now().Add(-24 * time.Hour)
	diff := updated.FinishesAt.Sub(expected)
	if diff > 5*time.Second || diff < -5*time.Second {
		t.Fatalf("expected finishes_at about 24h ago, got %v", updated.FinishesAt)
	}

	request = httptest.NewRequest(http.MethodPatch, "/api/v1/players/1/builds/9999/quick-finish", nil)
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodPatch, "/api/v1/players/1/builds/102/quick-finish", nil)
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", response.Code)
	}
}

func TestPlayerMailList(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	execAPITestSQLT(t, "INSERT INTO mails (receiver_id, title, body) VALUES ($1, $2, $3)", int64(1), "T", "B")
	mailID := queryAPITestInt64(t, "SELECT id FROM mails WHERE receiver_id = $1 ORDER BY id DESC LIMIT 1", int64(1))
	execAPITestSQLT(t, "INSERT INTO mail_attachments (mail_id, type, item_id, quantity) VALUES ($1, $2, $3, $4)", mailID, int64(2), int64(20001), int64(1))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players/1/mails", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
}

func TestPlayerGiveSkin(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	execAPITestSQLT(t, "INSERT INTO skins (id, name, ship_group) VALUES ($1, $2, $3)", int64(1), "Skin", int64(1))

	expectedExpiry := time.Date(2027, 1, 1, 9, 10, 0, 0, time.UTC)
	body := []byte(`{"skin_id":1,"expires_at":"2027-01-01T09:10:00Z"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/players/1/give-skin", bytes.NewBuffer(body))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/players/1/skins", nil)
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload playerSkinResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !payload.OK {
		t.Fatalf("expected ok response")
	}

	var found *types.PlayerSkinEntry
	for i := range payload.Data.Skins {
		skin := &payload.Data.Skins[i]
		if skin.SkinID == 1 {
			found = skin
			break
		}
	}
	if found == nil {
		t.Fatalf("expected skin to be granted")
	}
	if found.ExpiresAt == nil || !found.ExpiresAt.Equal(expectedExpiry) {
		t.Fatalf("expected expires_at to be %s", expectedExpiry.Format(time.RFC3339))
	}

}

func TestPlayerSearchWithMinLevel(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players/search?q=a&min_level=10&offset=0&limit=10", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
}

func TestPlayerBanDurationValidation(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	body := []byte(`{"permanent":true,"duration_sec":10}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/players/1/ban", bytes.NewBuffer(body))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestPlayerBanLiftTimestampValidation(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	body := []byte(`{"lift_timestamp":"not-a-time"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/players/1/ban", bytes.NewBuffer(body))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestPlayerBanDuration(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	body := []byte(`{"duration_sec":1}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/players/1/ban", bytes.NewBuffer(body))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	count := queryAPITestInt64(t, "SELECT COUNT(*) FROM punishments WHERE punished_id = $1 AND lift_timestamp IS NOT NULL", int64(1))
	if count == 0 {
		t.Fatalf("expected lift_timestamp to be set")
	}
}

func TestPlayerFilterBanned(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	execAPITestSQLT(t, "INSERT INTO punishments (punished_id, is_permanent) VALUES ($1, $2)", int64(2), true)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players?filter=banned&offset=0&limit=10", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
}

func TestPlayerFilterOnline(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players?filter=online&offset=0&limit=10", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
}

func TestPlayerFilterOnlineBanned(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	execAPITestSQLT(t, "INSERT INTO punishments (punished_id, is_permanent) VALUES ($1, $2)", int64(1), true)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players?filter=online,banned&offset=0&limit=10", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
}

func TestPlayerUpdateResourcesValidation(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	body := []byte(`{"resources":[]}`)
	request := httptest.NewRequest(http.MethodPut, "/api/v1/players/1/resources", bytes.NewBuffer(body))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestPlayerGiveItemValidation(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	body := []byte(`{"item_id":0,"amount":0}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/players/1/give-item", bytes.NewBuffer(body))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestPlayerGiveShipValidation(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	body := []byte(`{"ship_id":0}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/players/1/give-ship", bytes.NewBuffer(body))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestPlayerGiveSkinValidation(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	body := []byte(`{"skin_id":0}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/players/1/give-skin", bytes.NewBuffer(body))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestPlayerSendMailValidation(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	body := []byte(`{"title":"","body":""}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/players/1/send-mail", bytes.NewBuffer(body))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestPlayerKickNotOnline(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/players/1/kick", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", response.Code)
	}
}

func TestPlayerBanPermanentValidation(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	body := []byte(`{"permanent":true,"lift_timestamp":"2025-01-01T00:00:00Z"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/players/1/ban", bytes.NewBuffer(body))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestPlayerListLimitBounds(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/players?offset=0&limit=%d", 201), nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
}
