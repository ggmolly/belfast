package tests

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
	"github.com/ggmolly/belfast/internal/orm"
)

type playerListResponse struct {
	OK   bool                     `json:"ok"`
	Data types.PlayerListResponse `json:"data"`
}

type genericResponse struct {
	OK bool `json:"ok"`
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
	if testApp != nil {
		return
	}
	os.Setenv("MODE", "test")
	if ok := orm.InitDatabase(); !ok {
	}
	cfg := api.Config{Enabled: true, Port: 0}
	cfg.RuntimeConfig = &config.Config{}
	connection.BelfastInstance = connection.NewServer("127.0.0.1", 0, func(*[]byte, *connection.Client, int) {})
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

	item := orm.CommanderItem{CommanderID: 1, ItemID: 20001, Count: 1}
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

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players/1/skins", nil)
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
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

func TestPlayerBuildsList(t *testing.T) {
	setupTestAPI(t)
	seedPlayers(t)

	build := orm.Build{BuilderID: 1, ShipID: 1, FinishesAt: time.Now().Add(time.Hour)}
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

	body := []byte(`{"skin_id":1}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/players/1/give-skin", bytes.NewBuffer(body))
	response := httptest.NewRecorder()
	testApp.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
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
