package orm

import (
	"sync"
	"testing"

	"gorm.io/gorm"
)

var serverTestOnce sync.Once

func initServerTest(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	serverTestOnce.Do(func() {
		InitDatabase()
	})
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Server{}).Error; err != nil {
		t.Fatalf("clear servers: %v", err)
	}
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&ServerState{}).Error; err != nil {
		t.Fatalf("clear server states: %v", err)
	}
}

func TestServerCreate(t *testing.T) {
	initServerTest(t)

	serverState := ServerState{
		ID:          1,
		Description: "Online",
		Color:       "success",
	}

	if err := GormDB.Create(&serverState).Error; err != nil {
		t.Fatalf("create server state: %v", err)
	}

	server := Server{
		ID:      1,
		IP:      "127.0.0.1",
		Port:    8080,
		Name:    "Test Server",
		StateID: &serverState.ID,
	}

	if err := server.Create(); err != nil {
		t.Fatalf("create server: %v", err)
	}

	if server.ID != 1 {
		t.Fatalf("expected server id 1, got %d", server.ID)
	}
	if server.IP != "127.0.0.1" {
		t.Fatalf("expected ip '127.0.0.1', got %s", server.IP)
	}
	if server.Port != 8080 {
		t.Fatalf("expected port 8080, got %d", server.Port)
	}
	if server.Name != "Test Server" {
		t.Fatalf("expected name 'Test Server', got %s", server.Name)
	}
}

func TestServerUpdate(t *testing.T) {
	initServerTest(t)

	serverState := ServerState{
		ID:          1,
		Description: "Online",
		Color:       "success",
	}
	GormDB.Create(&serverState)

	server := Server{
		ID:      1,
		IP:      "127.0.0.1",
		Port:    8080,
		Name:    "Test Server",
		StateID: &serverState.ID,
	}
	server.Create()

	server.Name = "Updated Server"
	server.Port = 9090

	if err := server.Update(); err != nil {
		t.Fatalf("update server: %v", err)
	}

	var found Server
	if err := GormDB.First(&found, server.ID).Error; err != nil {
		t.Fatalf("find server: %v", err)
	}

	if found.Name != "Updated Server" {
		t.Fatalf("expected name 'Updated Server', got %s", found.Name)
	}
	if found.Port != 9090 {
		t.Fatalf("expected port 9090, got %d", found.Port)
	}
}

func TestServerRetrieve(t *testing.T) {
	initServerTest(t)

	serverState := ServerState{
		ID:          1,
		Description: "Online",
		Color:       "success",
	}
	GormDB.Create(&serverState)

	server := Server{
		ID:      1,
		IP:      "127.0.0.1",
		Port:    8080,
		Name:    "Test Server",
		StateID: &serverState.ID,
	}
	server.Create()

	t.Run("retrieve without greedy", func(t *testing.T) {
		var found Server
		found.ID = server.ID
		if err := found.Retrieve(false); err != nil {
			t.Fatalf("retrieve server: %v", err)
		}

		if found.IP != "127.0.0.1" {
			t.Fatalf("expected ip '127.0.0.1', got %s", found.IP)
		}
		if found.State.ID != 0 {
			t.Fatalf("expected state not to be loaded without greedy, got %d", found.State.ID)
		}
	})

	t.Run("retrieve with greedy", func(t *testing.T) {
		var found Server
		found.ID = server.ID
		if err := found.Retrieve(true); err != nil {
			t.Fatalf("retrieve server greedy: %v", err)
		}

		if found.IP != "127.0.0.1" {
			t.Fatalf("expected ip '127.0.0.1', got %s", found.IP)
		}
	})
}

func TestServerDelete(t *testing.T) {
	initServerTest(t)

	serverState := ServerState{
		ID:          1,
		Description: "Online",
		Color:       "success",
	}
	GormDB.Create(&serverState)

	server := Server{
		ID:      1,
		IP:      "127.0.0.1",
		Port:    8080,
		Name:    "Test Server",
		StateID: &serverState.ID,
	}
	server.Create()

	if err := server.Delete(); err != nil {
		t.Fatalf("delete server: %v", err)
	}

	var found Server
	err := GormDB.First(&found, server.ID).Error
	if err != gorm.ErrRecordNotFound {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestServerStateRelationships(t *testing.T) {
	initServerTest(t)

	serverState := ServerState{
		ID:          1,
		Description: "Online",
		Color:       "success",
	}
	GormDB.Create(&serverState)

	server1 := Server{
		ID:      1,
		IP:      "127.0.0.1",
		Port:    8080,
		Name:    "Server 1",
		StateID: &serverState.ID,
	}
	server2 := Server{
		ID:      2,
		IP:      "127.0.0.1",
		Port:    8081,
		Name:    "Server 2",
		StateID: &serverState.ID,
	}
	GormDB.Create(&server1)
	GormDB.Create(&server2)

	var foundState ServerState
	if err := GormDB.Preload("Servers").First(&foundState, serverState.ID).Error; err != nil {
		t.Fatalf("find server state with servers: %v", err)
	}

	if len(foundState.Servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(foundState.Servers))
	}
}

func TestServerOptionalFields(t *testing.T) {
	initServerTest(t)

	serverState := ServerState{
		ID:          1,
		Description: "Online",
		Color:       "success",
	}
	GormDB.Create(&serverState)

	server := Server{
		ID:      1,
		IP:      "127.0.0.1",
		Port:    8080,
		Name:    "Test Server",
		StateID: &serverState.ID,
	}

	server.Create()

	var found Server
	if err := GormDB.First(&found, server.ID).Error; err != nil {
		t.Fatalf("find server: %v", err)
	}

	if found.ProxyIP != nil {
		t.Fatalf("expected ProxyIP to be nil, got %v", found.ProxyIP)
	}
	if found.ProxyPort != nil {
		t.Fatalf("expected ProxyPort to be nil, got %v", found.ProxyPort)
	}

	proxyIP := "192.168.1.1"
	proxyPort := 9999
	server.ProxyIP = &proxyIP
	server.ProxyPort = &proxyPort
	server.Update()

	var updated Server
	if err := GormDB.First(&updated, server.ID).Error; err != nil {
		t.Fatalf("find updated server: %v", err)
	}

	if updated.ProxyIP == nil || *updated.ProxyIP != "192.168.1.1" {
		t.Fatalf("expected ProxyIP to be '192.168.1.1', got %v", updated.ProxyIP)
	}
	if updated.ProxyPort == nil || *updated.ProxyPort != 9999 {
		t.Fatalf("expected ProxyPort to be 9999, got %v", updated.ProxyPort)
	}
}
