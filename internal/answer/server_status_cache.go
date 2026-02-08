package answer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/logger"
)

const (
	serverStatusCacheTTL = 30 * time.Second
)

var (
	serverStatusCacheMu          sync.Mutex
	serverStatusCacheRefreshedAt time.Time
	serverStatusCacheEntries     map[uint32]serverStatusEntry
	serverStatusHTTPClient       = &http.Client{Timeout: 2 * time.Second}
)

type serverStatusPayload struct {
	OK   bool             `json:"ok"`
	Data serverStatusData `json:"data"`
}

type serverStatusData struct {
	Name        string `json:"name"`
	Commit      string `json:"commit"`
	Running     bool   `json:"running"`
	Accepting   bool   `json:"accepting"`
	Maintenance bool   `json:"maintenance"`
}

type serverStatusEntry struct {
	Name   string
	Commit string
	State  uint32
}

func getServerStatusCache(servers []config.ServerConfig) map[uint32]serverStatusEntry {
	serverStatusCacheMu.Lock()
	defer serverStatusCacheMu.Unlock()
	if time.Since(serverStatusCacheRefreshedAt) < serverStatusCacheTTL && serverStatusCacheEntries != nil {
		return serverStatusCacheEntries
	}
	entries := make(map[uint32]serverStatusEntry, len(servers))
	for i := range servers {
		server := servers[i]
		entries[server.ID] = resolveServerStatus(server)
	}
	serverStatusCacheEntries = entries
	serverStatusCacheRefreshedAt = time.Now().UTC()
	return serverStatusCacheEntries
}

func resolveServerStatus(server config.ServerConfig) serverStatusEntry {
	entry := serverStatusEntry{
		Name:  server.IP,
		State: SERVER_STATE_OFFLINE,
	}
	if server.AssertOnline {
		entry.State = SERVER_STATE_ONLINE
		return entry
	}
	if server.ApiPort == 0 {
		return entry
	}
	status, err := fetchServerStatus(server.IP, server.ApiPort)
	if err != nil {
		logger.LogEvent("Server", "StatusRefresh", fmt.Sprintf("status check failed for %s:%d: %s", server.IP, server.ApiPort, err.Error()), logger.LOG_LEVEL_WARN)
		return entry
	}
	name := strings.TrimSpace(status.Name)
	if name == "" {
		name = server.IP
	}
	entry.Name = name
	entry.Commit = shortCommit(status.Commit)
	if status.Maintenance {
		entry.State = SERVER_STATE_OFFLINE
		return entry
	}
	if status.Running && status.Accepting {
		entry.State = SERVER_STATE_ONLINE
		return entry
	}
	if status.Running && !status.Accepting {
		entry.State = SERVER_STATE_BUSY
		return entry
	}
	entry.State = SERVER_STATE_OFFLINE
	return entry
}

func fetchServerStatus(host string, apiPort int) (serverStatusData, error) {
	var status serverStatusData
	url := fmt.Sprintf("http://%s:%d/api/v1/server/status", host, apiPort)
	resp, err := serverStatusHTTPClient.Get(url)
	if err != nil {
		return status, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return status, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	var payload serverStatusPayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return status, err
	}
	if !payload.OK {
		return status, fmt.Errorf("status payload not ok")
	}
	return payload.Data, nil
}

func shortCommit(commit string) string {
	trimmed := strings.TrimSpace(commit)
	if trimmed == "" {
		return ""
	}
	if len(trimmed) > 7 {
		return trimmed[:7]
	}
	return trimmed
}
