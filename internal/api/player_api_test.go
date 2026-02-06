package api_test

import (
	"bytes"
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
	if err := orm.GormDB.Model(&orm.Punishment{}).Where("punished_id = ?", 1).Count(&count).Error; err != nil {
		t.Fatalf("failed to check punishment: %v", err)
	}
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

	var resource orm.OwnedResource
	if err := orm.GormDB.Where("commander_id = ? AND resource_id = ?", 1, 1).First(&resource).Error; err != nil {
		t.Fatalf("failed to fetch resource: %v", err)
	}
	if resource.Amount != 500 {
		t.Fatalf("expected amount 500, got %d", resource.Amount)
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

	var compensation orm.Compensation
	if err := orm.GormDB.Preload("Attachments").Where("commander_id = ?", 1).First(&compensation).Error; err != nil {
		t.Fatalf("failed to load compensation: %v", err)
	}

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
	if len(listPayload.Data.Compensations) != 1 {
		t.Fatalf("expected 1 compensation, got %d", len(listPayload.Data.Compensations))
	}

	request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/players/1/compensations/%d", compensation.ID), nil)
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	updateBody := `{"title":"Updated","attach_flag":true,"attachments":[{"type":1,"item_id":1,"quantity":5}]}`
	request = httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/players/1/compensations/%d", compensation.ID), bytes.NewBuffer([]byte(updateBody)))
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	if err := orm.GormDB.Preload("Attachments").First(&compensation, compensation.ID).Error; err != nil {
		t.Fatalf("failed to reload compensation: %v", err)
	}
	if compensation.Title != "Updated" {
		t.Fatalf("expected title updated, got %s", compensation.Title)
	}
	if !compensation.AttachFlag {
		t.Fatalf("expected attach flag true")
	}
	if len(compensation.Attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(compensation.Attachments))
	}

	request = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/players/1/compensations/%d", compensation.ID), nil)
	response = httptest.NewRecorder()
	testApp.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var count int64
	if err := orm.GormDB.Model(&orm.Compensation{}).Where("commander_id = ?", 1).Count(&count).Error; err != nil {
		t.Fatalf("failed to count compensations: %v", err)
	}
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

	var commander orm.Commander
	if err := orm.GormDB.Unscoped().First(&commander, 1).Error; err != nil {
		t.Fatalf("expected commander to exist: %v", err)
	}
	if commander.DeletedAt.Time.IsZero() {
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
	if err := orm.GormDB.Exec("DELETE FROM punishments").Error; err != nil {
		t.Fatalf("failed to clear punishments: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM owned_resources").Error; err != nil {
		t.Fatalf("failed to clear owned_resources: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM commander_items").Error; err != nil {
		t.Fatalf("failed to clear commander_items: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM commander_misc_items").Error; err != nil {
		t.Fatalf("failed to clear commander_misc_items: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM owned_ships").Error; err != nil {
		t.Fatalf("failed to clear owned_ships: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM builds").Error; err != nil {
		t.Fatalf("failed to clear builds: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM mails").Error; err != nil {
		t.Fatalf("failed to clear mails: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM mail_attachments").Error; err != nil {
		t.Fatalf("failed to clear mail_attachments: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM compensations").Error; err != nil {
		t.Fatalf("failed to clear compensations: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM compensation_attachments").Error; err != nil {
		t.Fatalf("failed to clear compensation_attachments: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM fleets").Error; err != nil {
		t.Fatalf("failed to clear fleets: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM owned_skins").Error; err != nil {
		t.Fatalf("failed to clear owned_skins: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM commanders").Error; err != nil {
		t.Fatalf("failed to clear commanders: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM resources").Error; err != nil {
		t.Fatalf("failed to clear resources: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM items").Error; err != nil {
		t.Fatalf("failed to clear items: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM ships").Error; err != nil {
		t.Fatalf("failed to clear ships: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM skins").Error; err != nil {
		t.Fatalf("failed to clear skins: %v", err)
	}

	seedDb()

	miscItem := orm.Item{ID: 30001, Name: "Misc Item", Rarity: 1, ShopID: 0, Type: 0, VirtualType: 0}
	if err := orm.GormDB.Create(&miscItem).Error; err != nil {
		t.Fatalf("failed to create misc item: %v", err)
	}

	commander1 := orm.Commander{
		CommanderID: 1,
		AccountID:   10,
		Name:        "Alpha",
		Level:       5,
		LastLogin:   time.Now().Add(-time.Hour),
	}
	commander2 := orm.Commander{
		CommanderID: 2,
		AccountID:   11,
		Name:        "Bravo",
		Level:       20,
		LastLogin:   time.Now().Add(-2 * time.Hour),
	}
	if err := orm.GormDB.Create(&commander1).Error; err != nil {
		t.Fatalf("failed to create commander1: %v", err)
	}
	if err := orm.GormDB.Create(&commander2).Error; err != nil {
		t.Fatalf("failed to create commander2: %v", err)
	}

	ship := orm.Ship{TemplateID: 1, Name: "Test Ship", RarityID: 2, Star: 1, Type: 1}
	if err := orm.GormDB.Create(&ship).Error; err != nil {
		t.Fatalf("failed to create ship: %v", err)
	}

	resource := orm.OwnedResource{CommanderID: 1, ResourceID: 1, Amount: 100}
	if err := orm.GormDB.Create(&resource).Error; err != nil {
		t.Fatalf("failed to create owned resource: %v", err)
	}
	resource2 := orm.OwnedResource{CommanderID: 1, ResourceID: 2, Amount: 30}
	if err := orm.GormDB.Create(&resource2).Error; err != nil {
		t.Fatalf("failed to create owned resource2: %v", err)
	}

	item := orm.CommanderItem{CommanderID: 1, ItemID: 20001, Count: 5}
	if err := orm.GormDB.Create(&item).Error; err != nil {
		t.Fatalf("failed to create owned item: %v", err)
	}

	misc := orm.CommanderMiscItem{CommanderID: 1, ItemID: 30001, Data: 2}
	if err := orm.GormDB.Create(&misc).Error; err != nil {
		t.Fatalf("failed to create misc item: %v", err)
	}
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

	skin1 := orm.Skin{ID: 1001, Name: "Test Skin A", ShipGroup: 1}
	skin2 := orm.Skin{ID: 1002, Name: "Test Skin B", ShipGroup: 1}
	if err := orm.GormDB.Create(&skin1).Error; err != nil {
		t.Fatalf("failed to create skin1: %v", err)
	}
	if err := orm.GormDB.Create(&skin2).Error; err != nil {
		t.Fatalf("failed to create skin2: %v", err)
	}
	expiresAt := time.Now().Add(24 * time.Hour)
	ownedSkin := orm.OwnedSkin{CommanderID: 1, SkinID: 1002, ExpiresAt: &expiresAt}
	if err := orm.GormDB.Create(&ownedSkin).Error; err != nil {
		t.Fatalf("failed to create owned skin: %v", err)
	}

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

	if err := orm.GormDB.Exec("DELETE FROM event_collections").Error; err != nil {
		t.Fatalf("failed to clear event_collections: %v", err)
	}

	owned := orm.OwnedShip{ID: 101, OwnerID: 1, ShipID: 1, Level: 1, Energy: 150}
	if err := orm.GormDB.Create(&owned).Error; err != nil {
		t.Fatalf("failed to create owned ship: %v", err)
	}
	busy := orm.EventCollection{CommanderID: 1, CollectionID: 1, StartTime: 1, FinishTime: 2, ShipIDs: orm.Int64List{int64(owned.ID)}}
	if err := orm.GormDB.Create(&busy).Error; err != nil {
		t.Fatalf("failed to create event collection: %v", err)
	}

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

	build := orm.Build{BuilderID: 1, ShipID: 1, PoolID: 1, FinishesAt: time.Now().Add(time.Hour)}
	if err := orm.GormDB.Create(&build).Error; err != nil {
		t.Fatalf("failed to create build: %v", err)
	}

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

	build1 := orm.Build{BuilderID: 1, ShipID: 1, PoolID: 1, FinishesAt: time.Now().Add(2 * time.Hour)}
	if err := orm.GormDB.Create(&build1).Error; err != nil {
		t.Fatalf("failed to create build1: %v", err)
	}
	build2 := orm.Build{BuilderID: 1, ShipID: 1, PoolID: 2, FinishesAt: time.Now().Add(-1 * time.Hour)}
	if err := orm.GormDB.Create(&build2).Error; err != nil {
		t.Fatalf("failed to create build2: %v", err)
	}

	if err := orm.GormDB.Model(&orm.Commander{}).Where("commander_id = ?", 1).
		Updates(map[string]interface{}{"draw_count1": 4, "draw_count10": 5, "exchange_count": 6}).Error; err != nil {
		t.Fatalf("failed to update counters: %v", err)
	}

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
	if payload.Data.WorklistList[0].PoolID != build1.PoolID {
		t.Fatalf("expected pool id %d, got %d", build1.PoolID, payload.Data.WorklistList[0].PoolID)
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

	var commander orm.Commander
	if err := orm.GormDB.First(&commander, 1).Error; err != nil {
		t.Fatalf("failed to reload commander: %v", err)
	}
	if commander.DrawCount1 != 2 || commander.DrawCount10 != 3 || commander.ExchangeCount != 7 {
		t.Fatalf("counters did not update")
	}
}

func TestUpdatePlayerBuild(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	ship2 := orm.Ship{TemplateID: 2, Name: "Build Ship", RarityID: 2, Star: 1, Type: 1}
	if err := orm.GormDB.Create(&ship2).Error; err != nil {
		t.Fatalf("failed to create ship2: %v", err)
	}

	build := orm.Build{ID: 99, BuilderID: 1, ShipID: 1, PoolID: 1, FinishesAt: time.Now().Add(2 * time.Hour)}
	if err := orm.GormDB.Create(&build).Error; err != nil {
		t.Fatalf("failed to create build: %v", err)
	}

	otherBuild := orm.Build{ID: 100, BuilderID: 2, ShipID: 1, PoolID: 1, FinishesAt: time.Now().Add(2 * time.Hour)}
	if err := orm.GormDB.Create(&otherBuild).Error; err != nil {
		t.Fatalf("failed to create other build: %v", err)
	}

	body := []byte(`{"ship_id":2}`)
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/players/1/builds/99", bytes.NewBuffer(body))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var updated orm.Build
	if err := orm.GormDB.First(&updated, 99).Error; err != nil {
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

	if err := orm.GormDB.First(&updated, 99).Error; err != nil {
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

	build := orm.Build{ID: 101, BuilderID: 1, ShipID: 1, PoolID: 1, FinishesAt: time.Now().Add(2 * time.Hour)}
	if err := orm.GormDB.Create(&build).Error; err != nil {
		t.Fatalf("failed to create build: %v", err)
	}

	otherBuild := orm.Build{ID: 102, BuilderID: 2, ShipID: 1, PoolID: 1, FinishesAt: time.Now().Add(2 * time.Hour)}
	if err := orm.GormDB.Create(&otherBuild).Error; err != nil {
		t.Fatalf("failed to create other build: %v", err)
	}

	request := httptest.NewRequest(http.MethodPatch, "/api/v1/players/1/builds/101/quick-finish", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var updated orm.Build
	if err := orm.GormDB.First(&updated, 101).Error; err != nil {
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

	mail := orm.Mail{ReceiverID: 1, Title: "T", Body: "B"}
	if err := orm.GormDB.Create(&mail).Error; err != nil {
		t.Fatalf("failed to create mail: %v", err)
	}
	attachment := orm.MailAttachment{MailID: mail.ID, Type: 2, ItemID: 20001, Quantity: 1}
	if err := orm.GormDB.Create(&attachment).Error; err != nil {
		t.Fatalf("failed to create mail attachment: %v", err)
	}

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

	skin := orm.Skin{ID: 1, Name: "Skin"}
	if err := orm.GormDB.Create(&skin).Error; err != nil {
		t.Fatalf("failed to create skin: %v", err)
	}

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

	var punishment orm.Punishment
	if err := orm.GormDB.Where("punished_id = ?", 1).First(&punishment).Error; err != nil {
		t.Fatalf("failed to load punishment: %v", err)
	}
	if punishment.LiftTimestamp == nil {
		t.Fatalf("expected lift_timestamp to be set")
	}
}

func TestPlayerFilterBanned(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	punishment := orm.Punishment{PunishedID: 2, IsPermanent: true}
	if err := orm.GormDB.Create(&punishment).Error; err != nil {
		t.Fatalf("failed to create punishment: %v", err)
	}

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

	punishment := orm.Punishment{PunishedID: 1, IsPermanent: true}
	if err := orm.GormDB.Create(&punishment).Error; err != nil {
		t.Fatalf("failed to create punishment: %v", err)
	}

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
