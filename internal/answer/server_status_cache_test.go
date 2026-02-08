package answer

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/api/response"
	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/config"
)

func TestServerStatusCacheMapping(t *testing.T) {
	onlineServer := httptest.NewServer(statusHandler(types.ServerStatusResponse{
		Name:        "Alpha",
		Commit:      "abcdef123456",
		Running:     true,
		Accepting:   true,
		Maintenance: false,
		UptimeSec:   1,
		UptimeHuman: "1s",
		ClientCount: 3,
	}))
	defer onlineServer.Close()

	busyServer := httptest.NewServer(statusHandler(types.ServerStatusResponse{
		Name:        "Beta",
		Commit:      "",
		Running:     true,
		Accepting:   false,
		Maintenance: false,
		UptimeSec:   1,
		UptimeHuman: "1s",
		ClientCount: 10,
	}))
	defer busyServer.Close()

	maintenanceServer := httptest.NewServer(statusHandler(types.ServerStatusResponse{
		Name:        "Gamma",
		Commit:      "1234567",
		Running:     true,
		Accepting:   true,
		Maintenance: true,
		UptimeSec:   1,
		UptimeHuman: "1s",
		ClientCount: 0,
	}))
	defer maintenanceServer.Close()

	onlineHost, onlinePort := parseTestServerHostPort(t, onlineServer.URL)
	busyHost, busyPort := parseTestServerHostPort(t, busyServer.URL)
	maintenanceHost, maintenancePort := parseTestServerHostPort(t, maintenanceServer.URL)

	servers := []config.ServerConfig{
		{ID: 1, IP: onlineHost, Port: 1001, ApiPort: onlinePort},
		{ID: 2, IP: busyHost, Port: 1002, ApiPort: busyPort},
		{ID: 3, IP: maintenanceHost, Port: 1003, ApiPort: maintenancePort},
		{ID: 4, IP: "192.0.2.1", Port: 1004, ApiPort: 0},
	}

	serverStatusCacheRefreshedAt = time.Time{}
	serverStatusCacheEntries = nil
	statuses := getServerStatusCache(servers)

	if status := statuses[1]; status.State != SERVER_STATE_ONLINE {
		t.Fatalf("expected online state, got %d", status.State)
	} else if status.Name != "Alpha" {
		t.Fatalf("expected name Alpha, got %q", status.Name)
	} else if status.Commit != "abcdef1" {
		t.Fatalf("expected short commit abcdef1, got %q", status.Commit)
	}

	if status := statuses[2]; status.State != SERVER_STATE_BUSY {
		t.Fatalf("expected busy state, got %d", status.State)
	} else if status.Name != "Beta" {
		t.Fatalf("expected name Beta, got %q", status.Name)
	}

	if status := statuses[3]; status.State != SERVER_STATE_OFFLINE {
		t.Fatalf("expected maintenance mapped to offline, got %d", status.State)
	}

	if status := statuses[4]; status.Name != "192.0.2.1" {
		t.Fatalf("expected fallback name to IP, got %q", status.Name)
	}

	serverInfo := buildServerInfo(servers, statuses)
	if got := serverInfo[0].GetName(); got != "Alpha (abcdef1)" {
		t.Fatalf("expected formatted name with commit, got %q", got)
	}
}

func TestServerStatusCacheAssertOnlineSkipsFetch(t *testing.T) {
	previousClient := serverStatusHTTPClient
	called := false
	serverStatusHTTPClient = &http.Client{
		Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			called = true
			return nil, errors.New("unexpected status fetch")
		}),
		Timeout: 50 * time.Millisecond,
	}
	defer func() { serverStatusHTTPClient = previousClient }()

	serverStatusCacheRefreshedAt = time.Time{}
	serverStatusCacheEntries = nil
	statuses := getServerStatusCache([]config.ServerConfig{{ID: 1, IP: "203.0.113.1", Port: 1001, ApiPort: 1234, AssertOnline: true}})
	if called {
		t.Fatalf("expected no status fetch when server assert_online is enabled")
	}
	if status := statuses[1]; status.State != SERVER_STATE_ONLINE {
		t.Fatalf("expected online state, got %d", status.State)
	}
}

func statusHandler(status types.ServerStatusResponse) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload := response.Success(status)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(payload)
	}
}

func parseTestServerHostPort(t *testing.T, rawURL string) (string, int) {
	t.Helper()
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	host, portString, err := net.SplitHostPort(parsed.Host)
	if err != nil {
		t.Fatalf("split host port: %v", err)
	}
	port, err := strconv.Atoi(portString)
	if err != nil {
		t.Fatalf("parse port: %v", err)
	}
	return host, port
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
