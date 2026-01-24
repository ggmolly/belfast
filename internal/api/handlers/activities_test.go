package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/orm"
)

var activityHandlerTestOnce sync.Once

func initActivityHandlerTestDB(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	activityHandlerTestOnce.Do(func() {
		orm.InitDatabase()
	})
}

func newActivityHandlerTestApp(t *testing.T) *iris.Application {
	initActivityHandlerTestDB(t)
	app := iris.New()
	handler := NewActivityHandler()
	RegisterActivityRoutes(app.Party("/api/v1/activities"), handler)
	if err := app.Build(); err != nil {
		t.Fatalf("build app: %v", err)
	}
	return app
}

func seedActivityTemplate(t *testing.T, id uint32) {
	t.Helper()
	key := strconv.FormatUint(uint64(id), 10)
	entry := orm.ConfigEntry{Category: "ShareCfg/activity_template.json", Key: key, Data: json.RawMessage(`{"id":` + key + `}`)}
	if err := orm.GormDB.Create(&entry).Error; err != nil {
		t.Fatalf("seed activity template: %v", err)
	}
}

func clearActivityAllowlist(t *testing.T) {
	t.Helper()
	if err := orm.GormDB.Where("category = ? AND key = ?", activityAllowlistCategory, activityAllowlistKey).Delete(&orm.ConfigEntry{}).Error; err != nil {
		t.Fatalf("clear allowlist: %v", err)
	}
}

func clearConfigEntries(t *testing.T) {
	t.Helper()
	if err := orm.GormDB.Exec("DELETE FROM config_entries").Error; err != nil {
		t.Fatalf("clear config entries: %v", err)
	}
}

func TestActivityAllowlistEndpoints(t *testing.T) {
	app := newActivityHandlerTestApp(t)
	clearConfigEntries(t)
	seedActivityTemplate(t, 1)
	seedActivityTemplate(t, 2)
	seedActivityTemplate(t, 3)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/activities/allowlist", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	var getResponse struct {
		OK   bool `json:"ok"`
		Data struct {
			IDs []uint32 `json:"ids"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&getResponse); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(getResponse.Data.IDs) != 0 {
		t.Fatalf("expected empty allowlist")
	}

	request = httptest.NewRequest(http.MethodPut, "/api/v1/activities/allowlist", strings.NewReader(`{"ids":[9999]}`))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodPut, "/api/v1/activities/allowlist", strings.NewReader(`{"ids":[2,1]}`))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	getResponse = struct {
		OK   bool `json:"ok"`
		Data struct {
			IDs []uint32 `json:"ids"`
		} `json:"data"`
	}{}
	if err := json.NewDecoder(response.Body).Decode(&getResponse); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(getResponse.Data.IDs) != 2 || getResponse.Data.IDs[0] != 1 || getResponse.Data.IDs[1] != 2 {
		t.Fatalf("expected allowlist [1 2], got %v", getResponse.Data.IDs)
	}

	request = httptest.NewRequest(http.MethodPatch, "/api/v1/activities/allowlist", strings.NewReader(`{"add":[3],"remove":[1]}`))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	getResponse = struct {
		OK   bool `json:"ok"`
		Data struct {
			IDs []uint32 `json:"ids"`
		} `json:"data"`
	}{}
	if err := json.NewDecoder(response.Body).Decode(&getResponse); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(getResponse.Data.IDs) != 2 || getResponse.Data.IDs[0] != 2 || getResponse.Data.IDs[1] != 3 {
		t.Fatalf("expected allowlist [2 3], got %v", getResponse.Data.IDs)
	}

	clearActivityAllowlist(t)
}
