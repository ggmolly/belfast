package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/orm"
)

var playerHandlerTestOnce sync.Once

func initPlayerHandlerTestDB(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	playerHandlerTestOnce.Do(func() {
		orm.InitDatabase()
	})
}

func newPlayerHandlerTestApp(t *testing.T) *iris.Application {
	initPlayerHandlerTestDB(t)
	app := iris.New()
	handler := NewPlayerHandler()
	RegisterPlayerRoutes(app.Party("/api/v1/players"), handler)
	if err := app.Build(); err != nil {
		t.Fatalf("build app: %v", err)
	}
	return app
}

func TestPlayerBuffEndpoints(t *testing.T) {
	app := newPlayerHandlerTestApp(t)
	commanderID := uint32(9100)
	if err := orm.GormDB.Where("commander_id = ?", commanderID).Delete(&orm.CommanderBuff{}).Error; err != nil {
		t.Fatalf("clear commander buffs: %v", err)
	}
	if err := orm.GormDB.Unscoped().Where("commander_id = ?", commanderID).Delete(&orm.Commander{}).Error; err != nil {
		t.Fatalf("clear commander: %v", err)
	}

	commander := orm.Commander{
		CommanderID: commanderID,
		AccountID:   1,
		Level:       1,
		Exp:         0,
		Name:        "Test Commander",
		LastLogin:   time.Now().UTC(),
		RoomID:      0,
	}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}

	now := time.Now().UTC()
	expiresAt := now.Add(24 * time.Hour)
	expiredAt := now.Add(-24 * time.Hour)
	payload := strings.NewReader("{\"buff_id\": 100, \"expires_at\": \"" + expiresAt.Format(time.RFC3339) + "\"}")
	request := httptest.NewRequest(http.MethodPost, "/api/v1/players/9100/buffs", payload)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if err := orm.UpsertCommanderBuff(commanderID, 200, expiredAt); err != nil {
		t.Fatalf("create expired buff: %v", err)
	}

	var created orm.CommanderBuff
	if err := orm.GormDB.First(&created, "commander_id = ? AND buff_id = ?", commanderID, 100).Error; err != nil {
		t.Fatalf("load buff entry: %v", err)
	}
	if created.ExpiresAt.UTC().Format(time.RFC3339) != expiresAt.Format(time.RFC3339) {
		t.Fatalf("expected expires_at %s, got %s", expiresAt.Format(time.RFC3339), created.ExpiresAt.UTC().Format(time.RFC3339))
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/players/9100/buffs", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	var listResponse struct {
		OK   bool `json:"ok"`
		Data struct {
			Buffs []types.PlayerBuffEntry `json:"buffs"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&listResponse); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !listResponse.OK {
		t.Fatalf("expected ok response")
	}
	if len(listResponse.Data.Buffs) != 2 {
		t.Fatalf("expected 2 buffs, got %d", len(listResponse.Data.Buffs))
	}

	request = httptest.NewRequest(http.MethodDelete, "/api/v1/players/9100/buffs/200", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/players/9100/buffs?active=true", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	listResponse = struct {
		OK   bool `json:"ok"`
		Data struct {
			Buffs []types.PlayerBuffEntry `json:"buffs"`
		} `json:"data"`
	}{}
	if err := json.NewDecoder(response.Body).Decode(&listResponse); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(listResponse.Data.Buffs) != 1 {
		t.Fatalf("expected 1 active buff, got %d", len(listResponse.Data.Buffs))
	}
	if listResponse.Data.Buffs[0].BuffID != 100 {
		t.Fatalf("expected buff id 100, got %d", listResponse.Data.Buffs[0].BuffID)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/players/9100/buffs", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	listResponse = struct {
		OK   bool `json:"ok"`
		Data struct {
			Buffs []types.PlayerBuffEntry `json:"buffs"`
		} `json:"data"`
	}{}
	if err := json.NewDecoder(response.Body).Decode(&listResponse); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(listResponse.Data.Buffs) != 1 {
		t.Fatalf("expected 1 buff after delete, got %d", len(listResponse.Data.Buffs))
	}
}
