package answer

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func loadLocalLoginConfig(t *testing.T, skipOnboarding bool) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	content := fmt.Sprintf(`[belfast]
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

[create_player]
skip_onboarding = %t
name_blacklist = []
name_illegal_pattern = ""
`, skipOnboarding)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if _, err := config.Load(path); err != nil {
		t.Fatalf("load config: %v", err)
	}
}

func setupLocalLoginTest(t *testing.T, skipOnboarding bool) *connection.Client {
	t.Helper()
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	loadLocalLoginConfig(t, skipOnboarding)
	clearTable(t, &orm.LocalAccount{})
	clearTable(t, &orm.YostarusMap{})
	clearTable(t, &orm.Commander{})
	clearTable(t, &orm.OwnedShip{})
	clearTable(t, &orm.CommanderItem{})
	clearTable(t, &orm.OwnedResource{})
	return &connection.Client{}
}

func decodeLoginResponse(t *testing.T, client *connection.Client, expectedID int, message proto.Message) {
	t.Helper()
	buffer := client.Buffer.Bytes()
	if len(buffer) == 0 {
		t.Fatalf("expected response buffer")
	}
	packetID := packets.GetPacketId(0, &buffer)
	if packetID != expectedID {
		t.Fatalf("expected packet %d, got %d", expectedID, packetID)
	}
	packetSize := packets.GetPacketSize(0, &buffer) + 2
	if len(buffer) < packetSize {
		t.Fatalf("expected packet size %d, got %d", packetSize, len(buffer))
	}
	payloadStart := packets.HEADER_SIZE
	payloadEnd := payloadStart + (packetSize - packets.HEADER_SIZE)
	if err := proto.Unmarshal(buffer[payloadStart:payloadEnd], message); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	client.Buffer.Reset()
}

func TestLocalLoginSuccessNoCommander(t *testing.T) {
	client := setupLocalLoginTest(t, false)
	if err := orm.GormDB.Create(&orm.LocalAccount{Arg2: 900020, Account: "local", Password: "pass", MailBox: ""}).Error; err != nil {
		t.Fatalf("seed local account: %v", err)
	}
	payload := &protobuf.CS_10020{
		LoginType: proto.Uint32(2),
		Arg1:      proto.String("local"),
		Arg2:      proto.String("pass"),
		CheckKey:  proto.String("check"),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := Forge_SC10021(&buf, client); err != nil {
		t.Fatalf("Forge_SC10021 failed: %v", err)
	}
	response := &protobuf.SC_10021{}
	decodeLoginResponse(t, client, 10021, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if response.GetAccountId() != 0 {
		t.Fatalf("expected account id 0, got %d", response.GetAccountId())
	}
	if response.GetServerTicket() != formatServerTicket(900020) {
		t.Fatalf("unexpected server ticket %s", response.GetServerTicket())
	}
}

func TestLocalLoginSuccessWithCommander(t *testing.T) {
	client := setupLocalLoginTest(t, false)
	if err := orm.GormDB.Create(&orm.LocalAccount{Arg2: 900021, Account: "local", Password: "pass", MailBox: ""}).Error; err != nil {
		t.Fatalf("seed local account: %v", err)
	}
	if err := orm.GormDB.Create(&orm.YostarusMap{Arg2: 900021, AccountID: 910000}).Error; err != nil {
		t.Fatalf("seed yostarus map: %v", err)
	}
	payload := &protobuf.CS_10020{
		LoginType: proto.Uint32(2),
		Arg1:      proto.String("local"),
		Arg2:      proto.String("pass"),
		CheckKey:  proto.String("check"),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := Forge_SC10021(&buf, client); err != nil {
		t.Fatalf("Forge_SC10021 failed: %v", err)
	}
	response := &protobuf.SC_10021{}
	decodeLoginResponse(t, client, 10021, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if response.GetAccountId() != 910000 {
		t.Fatalf("expected account id 910000, got %d", response.GetAccountId())
	}
}

func TestLocalLoginWrongPassword(t *testing.T) {
	client := setupLocalLoginTest(t, false)
	if err := orm.GormDB.Create(&orm.LocalAccount{Arg2: 900022, Account: "local", Password: "pass", MailBox: ""}).Error; err != nil {
		t.Fatalf("seed local account: %v", err)
	}
	payload := &protobuf.CS_10020{
		LoginType: proto.Uint32(2),
		Arg1:      proto.String("local"),
		Arg2:      proto.String("wrong"),
		CheckKey:  proto.String("check"),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := Forge_SC10021(&buf, client); err != nil {
		t.Fatalf("Forge_SC10021 failed: %v", err)
	}
	response := &protobuf.SC_10021{}
	decodeLoginResponse(t, client, 10021, response)
	if response.GetResult() != 1020 {
		t.Fatalf("expected result 1020, got %d", response.GetResult())
	}
}

func TestLocalLoginUnknownAccount(t *testing.T) {
	client := setupLocalLoginTest(t, false)
	payload := &protobuf.CS_10020{
		LoginType: proto.Uint32(2),
		Arg1:      proto.String("missing"),
		Arg2:      proto.String("pass"),
		CheckKey:  proto.String("check"),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := Forge_SC10021(&buf, client); err != nil {
		t.Fatalf("Forge_SC10021 failed: %v", err)
	}
	response := &protobuf.SC_10021{}
	decodeLoginResponse(t, client, 10021, response)
	if response.GetResult() != 1010 {
		t.Fatalf("expected result 1010, got %d", response.GetResult())
	}
}

func TestLocalLoginSkipOnboardingCreatesCommander(t *testing.T) {
	client := setupLocalLoginTest(t, true)
	if err := orm.GormDB.Create(&orm.LocalAccount{Arg2: 900023, Account: "local", Password: "pass", MailBox: ""}).Error; err != nil {
		t.Fatalf("seed local account: %v", err)
	}
	payload := &protobuf.CS_10020{
		LoginType: proto.Uint32(2),
		Arg1:      proto.String("local"),
		Arg2:      proto.String("pass"),
		CheckKey:  proto.String("check"),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := Forge_SC10021(&buf, client); err != nil {
		t.Fatalf("Forge_SC10021 failed: %v", err)
	}
	response := &protobuf.SC_10021{}
	decodeLoginResponse(t, client, 10021, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if response.GetAccountId() == 0 {
		t.Fatalf("expected non-zero account id")
	}
	var mapping orm.YostarusMap
	if err := orm.GormDB.Where("arg2 = ?", 900023).First(&mapping).Error; err != nil {
		t.Fatalf("fetch yostarus map: %v", err)
	}
	if mapping.AccountID != response.GetAccountId() {
		t.Fatalf("expected mapping account %d, got %d", response.GetAccountId(), mapping.AccountID)
	}
}
