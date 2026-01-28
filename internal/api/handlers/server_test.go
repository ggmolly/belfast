package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kataras/iris/v12"

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
