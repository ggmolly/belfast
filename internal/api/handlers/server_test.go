package handlers

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/middleware"
	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
)

func TestServerGetConfig(t *testing.T) {
	t.Setenv("MODE", "test")
	orm.InitDatabase()

	app := iris.New()
	cfg := config.Config{
		Belfast: config.BelfastConfig{
			BindAddress: "127.0.0.1",
			Port:        8080,
			Maintenance: false,
		},
		API: config.APIConfig{
			Enabled:     true,
			Port:        8081,
			Environment: "test",
			CORSOrigins: []string{"http://localhost:8080"},
		},
		Region: config.RegionConfig{
			Default: "EN",
		},
	}
	cfg.Auth.DisableAuth = true
	app.UseRouter(middleware.Auth(&cfg))
	handler := &ServerHandler{Config: &cfg}
	RegisterServerRoutes(app.Party("/api/v1/server"), handler)

	if err := app.Build(); err != nil {
		t.Fatalf("build app: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/server/config", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}

	var result struct {
		Ok bool `json:"ok"`
	}

	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if !result.Ok {
		t.Fatalf("expected ok true, got %v. Response: %s", result.Ok, response.Body.String())
	}
}

func TestServerStatusIncludesMaintenance(t *testing.T) {
	t.Setenv("MODE", "test")
	orm.InitDatabase()

	app := iris.New()
	cfg := config.Config{
		Belfast: config.BelfastConfig{
			BindAddress: "127.0.0.1",
			Port:        8080,
			Maintenance: false,
		},
		API: config.APIConfig{
			Enabled:     true,
			Port:        8081,
			Environment: "test",
			CORSOrigins: []string{"http://localhost:8080"},
		},
		Region: config.RegionConfig{
			Default: "EN",
		},
	}
	cfg.Auth.DisableAuth = true
	app.UseRouter(middleware.Auth(&cfg))
	connection.NewServer("127.0.0.1", 8080, func(pkt *[]byte, c *connection.Client, size int) {})
	connection.BelfastInstance.SetMaintenance(true)
	handler := &ServerHandler{Config: &cfg}
	RegisterServerRoutes(app.Party("/api/v1/server"), handler)
	if err := app.Build(); err != nil {
		t.Fatalf("build app: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/server/status", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}

	var result struct {
		Ok   bool `json:"ok"`
		Data struct {
			Maintenance bool `json:"maintenance"`
		} `json:"data"`
	}

	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !result.Ok {
		t.Fatalf("expected ok true, got %v. Response: %s", result.Ok, response.Body.String())
	}
	if !result.Data.Maintenance {
		t.Fatalf("expected maintenance true")
	}
}

func TestServerUpdateConfig(t *testing.T) {
	t.Setenv("MODE", "test")
	orm.InitDatabase()

	app := iris.New()
	cfg := config.Config{
		Belfast: config.BelfastConfig{
			BindAddress: "127.0.0.1",
			Port:        8080,
			Maintenance: false,
		},
		API: config.APIConfig{
			Enabled:     true,
			Port:        8081,
			Environment: "test",
			CORSOrigins: []string{"http://localhost:8080"},
		},
		Region: config.RegionConfig{
			Default: "EN",
		},
	}
	cfg.Auth.DisableAuth = true
	app.UseRouter(middleware.Auth(&cfg))
	handler := &ServerHandler{Config: &cfg}
	RegisterServerRoutes(app.Party("/api/v1/server"), handler)

	if err := app.Build(); err != nil {
		t.Fatalf("build app: %v", err)
	}

	reqBody := struct {
		BindAddress string `json:"bind_address"`
		Port        int    `json:"port"`
		Region      string `json:"region"`
	}{
		BindAddress: "0.0.0.0",
		Port:        9090,
		Region:      "CN",
	}

	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodPut, "/api/v1/server/config", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}

	var result struct {
		Ok bool `json:"ok"`
	}

	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if !result.Ok {
		t.Fatalf("expected ok true, got %v. Response: %s", result.Ok, response.Body.String())
	}
}

func TestServerUpdateConfigInvalidRegion(t *testing.T) {
	t.Setenv("MODE", "test")
	orm.InitDatabase()

	app := iris.New()
	cfg := config.Config{
		Belfast: config.BelfastConfig{
			BindAddress: "127.0.0.1",
			Port:        8080,
			Maintenance: false,
		},
		API: config.APIConfig{
			Enabled:     true,
			Port:        8081,
			Environment: "test",
			CORSOrigins: []string{"http://localhost:8080"},
		},
		Region: config.RegionConfig{
			Default: "EN",
		},
	}
	cfg.Auth.DisableAuth = true
	app.UseRouter(middleware.Auth(&cfg))
	handler := &ServerHandler{Config: &cfg}
	RegisterServerRoutes(app.Party("/api/v1/server"), handler)

	if err := app.Build(); err != nil {
		t.Fatalf("build app: %v", err)
	}

	reqBody := struct {
		BindAddress string `json:"bind_address"`
		Port        int    `json:"port"`
		Region      string `json:"region"`
	}{
		BindAddress: "0.0.0.0",
		Port:        9090,
		Region:      "INVALID",
	}

	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodPut, "/api/v1/server/config", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}

	var result struct {
		Ok      bool   `json:"ok"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if result.Ok {
		t.Fatalf("expected ok false for invalid region, got true")
	}
}

func TestServerUpdateConfigInvalidJSON(t *testing.T) {
	t.Setenv("MODE", "test")
	orm.InitDatabase()

	app := iris.New()
	cfg := config.Config{
		Belfast: config.BelfastConfig{
			BindAddress: "127.0.0.1",
			Port:        8080,
			Maintenance: false,
		},
		API: config.APIConfig{
			Enabled:     true,
			Port:        8081,
			Environment: "test",
			CORSOrigins: []string{"http://localhost:8080"},
		},
		Region: config.RegionConfig{
			Default: "EN",
		},
	}
	cfg.Auth.DisableAuth = true
	app.UseRouter(middleware.Auth(&cfg))
	handler := &ServerHandler{Config: &cfg}
	RegisterServerRoutes(app.Party("/api/v1/server"), handler)

	if err := app.Build(); err != nil {
		t.Fatalf("build app: %v", err)
	}

	request := httptest.NewRequest(http.MethodPut, "/api/v1/server/config", strings.NewReader("{invalid json"))
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}

	var result struct {
		Ok bool `json:"ok"`
	}

	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if result.Ok {
		t.Fatalf("expected ok false for invalid json, got true")
	}
}

func newServerTestApp(t *testing.T) *iris.Application {
	t.Setenv("MODE", "test")
	orm.InitDatabase()
	configPath := t.TempDir() + "/config.toml"
	if err := os.WriteFile(configPath, []byte("[belfast]\nmaintenance = false\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	cfg := config.Config{
		Belfast: config.BelfastConfig{
			BindAddress: "127.0.0.1",
			Port:        8080,
			Maintenance: false,
		},
		API: config.APIConfig{
			Enabled:     true,
			Port:        8081,
			Environment: "test",
			CORSOrigins: []string{"http://localhost:8080"},
		},
		Region: config.RegionConfig{
			Default: "EN",
		},
		Path: configPath,
	}
	connection.NewServer("127.0.0.1", 8080, func(pkt *[]byte, c *connection.Client, size int) {})
	app := iris.New()
	cfg.Auth.DisableAuth = true
	app.UseRouter(middleware.Auth(&cfg))
	handler := &ServerHandler{Config: &cfg}
	RegisterServerRoutes(app.Party("/api/v1/server"), handler)
	if err := app.Build(); err != nil {
		t.Fatalf("build app: %v", err)
	}
	return app
}

func TestServerStartStopRestartMaintenance(t *testing.T) {
	app := newServerTestApp(t)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/server/start", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodPost, "/api/v1/server/stop", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodPost, "/api/v1/server/restart", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodPost, "/api/v1/server/maintenance", strings.NewReader("{invalid"))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodPost, "/api/v1/server/maintenance", strings.NewReader(`{"enabled":true}`))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/server/maintenance", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
}

func TestServerStatsMetricsConnections(t *testing.T) {
	app := newServerTestApp(t)
	server := connection.BelfastInstance
	client := &connection.Client{
		IP:          net.ParseIP("127.0.0.1"),
		Port:        12345,
		Hash:        123,
		ConnectedAt: time.Now().UTC(),
	}
	server.AddClient(client)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/server/stats", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/server/metrics", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/server/connections", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/server/connections/invalid", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/server/connections/999", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/server/connections/123", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodDelete, "/api/v1/server/connections/123", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodDelete, "/api/v1/server/connections/999", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", response.Code)
	}
}

func TestServerUptime(t *testing.T) {
	app := newServerTestApp(t)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/server/uptime", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
}

func TestServerHelpers(t *testing.T) {
	if _, err := parseHash("bad"); err == nil {
		t.Fatalf("expected parseHash error")
	}
	if _, err := parseHash("123"); err != nil {
		t.Fatalf("expected parseHash success")
	}
	if commanderID(&connection.Client{}) != 0 {
		t.Fatalf("expected commander id 0 for nil commander")
	}
	if totals := aggregateMetrics([]*connection.Client{{}, {}}); totals.QueueMax != 0 {
		t.Fatalf("expected zero queue max")
	}
}
