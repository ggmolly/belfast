package answer_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/api"
	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/connection"
)

func TestConfigLoadMaintenance(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	content := `[belfast]
bind_address = "127.0.0.1"
port = 8080
maintenance = true

[api]
enabled = false
port = 9999
environment = "test"
cors_origins = ["*"]

[database]
path = "data/test.db"

[region]
default = "EN"
`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if !cfg.Belfast.Maintenance {
		t.Fatalf("expected maintenance enabled")
	}
	if cfg.Belfast.Port != 8080 {
		t.Fatalf("expected port 8080, got %d", cfg.Belfast.Port)
	}
}

func TestPersistMaintenanceUpdatesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	content := `[belfast]
bind_address = "127.0.0.1"
port = 8080
maintenance = false

[api]
enabled = true
port = 9999
environment = "test"
cors_origins = ["*"]

[auth]
disable_auth = true
`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if err := cfg.PersistMaintenance(true); err != nil {
		t.Fatalf("failed to persist maintenance: %v", err)
	}
	if !cfg.Belfast.Maintenance {
		t.Fatalf("expected maintenance enabled")
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	if !strings.Contains(string(updated), "maintenance = true") {
		t.Fatalf("expected maintenance flag to be true")
	}
}

func TestSetMaintenanceDisconnectsClients(t *testing.T) {
	server := connection.NewServer("127.0.0.1", 80, func(*[]byte, *connection.Client, int) {})
	conn1, conn2 := net.Pipe()
	defer conn2.Close()
	client := &connection.Client{
		IP:         net.IPv4(127, 0, 0, 1),
		Port:       1000,
		Connection: &conn1,
		Server:     server,
		Hash:       1,
	}
	server.AddClient(client)
	done := make(chan struct{})
	go func() {
		_, _ = io.ReadAll(conn2)
		close(done)
	}()
	server.SetMaintenance(true)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("expected connection to close")
	}
	if !server.MaintenanceEnabled() {
		t.Fatalf("expected maintenance enabled")
	}
	if !client.IsClosed() {
		t.Fatalf("expected client closed")
	}
	serverClients := server.ClientCount()
	if serverClients != 0 {
		t.Fatalf("expected no remaining clients, got %d", serverClients)
	}
}

func TestServerMaintenanceEndpoint(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	content := `[belfast]
bind_address = "127.0.0.1"
port = 8080
maintenance = false

[api]
enabled = true
port = 9999
environment = "test"
cors_origins = ["*"]

[auth]
disable_auth = true
`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	server := connection.NewServer("127.0.0.1", 80, func(*[]byte, *connection.Client, int) {})
	appConfig := api.LoadConfig(cfg)
	app := api.NewApp(appConfig)
	app.Build()
	payload := []byte(`{"enabled":true}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/server/maintenance", bytes.NewReader(payload))
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	var body struct {
		OK   bool                            `json:"ok"`
		Data types.ServerMaintenanceResponse `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !body.OK || !body.Data.Enabled {
		t.Fatalf("expected maintenance enabled response")
	}
	if !server.MaintenanceEnabled() {
		t.Fatalf("expected maintenance enabled")
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	if !strings.Contains(string(updated), "maintenance = true") {
		t.Fatalf("expected maintenance flag to be true")
	}
}
