package answer_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func loadTestConfig(t *testing.T) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	content := `[belfast]
bind_address = "127.0.0.1"
port = 80
maintenance = false

[api]
enabled = false
port = 0
environment = "test"
cors_origins = []

[database]
path = "data/test.db"

[region]
default = "EN"
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	if _, err := config.Load(path); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
}

func TestJoinServerNoKickWhenNoExistingSession(t *testing.T) {
	loadTestConfig(t)
	if err := orm.GormDB.Create(&orm.Commander{
		AccountID:   900001,
		CommanderID: 900001,
		Name:        "Test Commander 900001",
	}).Error; err != nil {
		t.Fatalf("failed to seed commander: %v", err)
	}

	server := connection.NewServer("127.0.0.1", 8080, func(*[]byte, *connection.Client, int) {})
	client := &connection.Client{Server: server}
	payload := &protobuf.CS_10022{
		AccountId:    proto.Uint32(900001),
		ServerTicket: proto.String("=*=*=*=BELFAST=*=*=*="),
		Platform:     proto.String("0"),
		Serverid:     proto.Uint32(1),
		CheckKey:     proto.String("check"),
		DeviceId:     proto.String("device-100"),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	if _, _, err := answer.JoinServer(&buf, client); err != nil {
		t.Fatalf("JoinServer failed: %v", err)
	}

	if client.Commander.CommanderID != 900001 {
		t.Fatalf("expected commander ID 900001, got %d", client.Commander.CommanderID)
	}

	if client.IsClosed() {
		t.Fatalf("expected client to not be closed when no existing session")
	}
}

func TestJoinServerKicksExistingSession(t *testing.T) {
	loadTestConfig(t)
	if err := orm.GormDB.Create(&orm.Commander{
		AccountID:   900002,
		CommanderID: 900002,
		Name:        "Test Commander 900002",
	}).Error; err != nil {
		t.Fatalf("failed to seed commander: %v", err)
	}

	server := connection.NewServer("127.0.0.1", 8080, func(*[]byte, *connection.Client, int) {})

	existingClient := &connection.Client{Server: server}
	existingClient.Commander = &orm.Commander{CommanderID: 900002}
	existingClient.Commander.Load()
	existingClient.Hash = 2001
	server.AddClient(existingClient)

	payload := &protobuf.CS_10022{
		AccountId:    proto.Uint32(900002),
		ServerTicket: proto.String("=*=*=*=BELFAST=*=*=*="),
		Platform:     proto.String("0"),
		Serverid:     proto.Uint32(1),
		CheckKey:     proto.String("check"),
		DeviceId:     proto.String("device-200"),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	newClient := &connection.Client{Server: server}
	if _, _, err := answer.JoinServer(&buf, newClient); err != nil {
		t.Fatalf("JoinServer failed: %v", err)
	}

	if newClient.Commander.CommanderID != 900002 {
		t.Fatalf("expected commander ID 900002, got %d", newClient.Commander.CommanderID)
	}

	if newClient.IsClosed() {
		t.Fatalf("expected new client to not be closed")
	}

	if !existingClient.IsClosed() {
		t.Fatalf("expected existing client to be closed")
	}

	_, found := server.FindClient(existingClient.Hash)
	if found {
		t.Fatalf("expected existing client to be removed from server")
	}

	buffer := existingClient.Buffer.Bytes()
	if len(buffer) == 0 {
		t.Fatalf("expected disconnect packet in kicked client buffer")
	}
	packetID := int(buffer[3])<<8 | int(buffer[4])
	if packetID != 10999 {
		t.Fatalf("expected packet 10999, got %d", packetID)
	}
}

func TestJoinServerMultipleDeviceSwaps(t *testing.T) {
	loadTestConfig(t)
	if err := orm.GormDB.Create(&orm.Commander{
		AccountID:   900003,
		CommanderID: 900003,
		Name:        "Test Commander 900003",
	}).Error; err != nil {
		t.Fatalf("failed to seed commander: %v", err)
	}

	server := connection.NewServer("127.0.0.1", 8080, func(*[]byte, *connection.Client, int) {})

	payloadTemplate := &protobuf.CS_10022{
		AccountId:    proto.Uint32(900003),
		ServerTicket: proto.String("=*=*=*=BELFAST=*=*=*="),
		Platform:     proto.String("0"),
		Serverid:     proto.Uint32(1),
		CheckKey:     proto.String("check"),
	}

	var clients []*connection.Client
	for i := 0; i < 5; i++ {
		deviceID := fmt.Sprintf("device-%d", i)
		payload := *payloadTemplate
		payload.DeviceId = proto.String(deviceID)
		buf, err := proto.Marshal(&payload)
		if err != nil {
			t.Fatalf("failed to marshal payload for client %d: %v", i, err)
		}

		client := &connection.Client{Server: server}
		if _, _, err := answer.JoinServer(&buf, client); err != nil {
			t.Fatalf("JoinServer failed for client %d: %v", i, err)
		}

		client.Hash = uint32(3000 + i)
		server.AddClient(client)
		clients = append(clients, client)

		if i > 0 {
			if !clients[i-1].IsClosed() {
				t.Fatalf("expected previous client %d to be closed", i-1)
			}
		}
	}

	for i, client := range clients {
		if i == len(clients)-1 {
			if client.IsClosed() {
				t.Fatalf("expected last client to not be closed")
			}
			if client.Commander.CommanderID != 900003 {
				t.Fatalf("expected last client to have commander 900003")
			}
		} else {
			if !client.IsClosed() {
				t.Fatalf("expected client %d to be closed", i)
			}
		}
	}

	if server.ClientCount() != 1 {
		t.Fatalf("expected 1 active client, got %d", server.ClientCount())
	}
}
