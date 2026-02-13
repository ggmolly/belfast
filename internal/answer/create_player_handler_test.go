package answer_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

const serverTicketPrefix = "=*=*=*=BELFAST=*=*=*="

func decodeResponsePacket(t *testing.T, client *connection.Client, expectedID int, message proto.Message) {
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

func formatTomlStringList(values []string) string {
	if len(values) == 0 {
		return "[]"
	}
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, fmt.Sprintf("%q", value))
	}
	return fmt.Sprintf("[%s]", strings.Join(quoted, ", "))
}

func loadCreatePlayerConfig(t *testing.T, skipOnboarding bool, blacklist []string, pattern string) {
	t.Helper()
	resetCreatePlayerTables(t)
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
name_blacklist = %s
name_illegal_pattern = %q
`, skipOnboarding, formatTomlStringList(blacklist), pattern)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	if _, err := config.Load(path); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
}

func resetCreatePlayerTables(t *testing.T) {
	t.Helper()
	execAnswerExternalTestSQLT(t, "DELETE FROM commanders")
	execAnswerExternalTestSQLT(t, "DELETE FROM device_auth_maps")
	execAnswerExternalTestSQLT(t, "DELETE FROM yostarus_maps")
}

func TestCreateNewPlayerSuccess(t *testing.T) {
	loadCreatePlayerConfig(t, false, nil, "")
	client := &connection.Client{AuthArg2: 900001}
	payload := &protobuf.CS_10024{
		NickName: proto.String("Molly"),
		ShipId:   proto.Uint32(201211),
		DeviceId: proto.String("device-100"),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.CreateNewPlayer(&buf, client); err != nil {
		t.Fatalf("CreateNewPlayer failed: %v", err)
	}
	response := &protobuf.SC_10025{}
	decodeResponsePacket(t, client, 10025, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if response.GetUserId() == 0 {
		t.Fatalf("expected non-zero user id")
	}
	deviceMapping, err := orm.GetDeviceAuthMapByDeviceID("device-100")
	if err != nil {
		t.Fatalf("failed to fetch device mapping: %v", err)
	}
	if deviceMapping.AccountID != response.GetUserId() {
		t.Fatalf("expected device mapping account %d, got %d", response.GetUserId(), deviceMapping.AccountID)
	}
	yostarus, err := orm.GetYostarusMapByArg2(client.AuthArg2)
	if err != nil {
		t.Fatalf("failed to fetch yostarus map: %v", err)
	}
	if yostarus.AccountID != response.GetUserId() {
		t.Fatalf("expected yostarus account %d, got %d", response.GetUserId(), yostarus.AccountID)
	}
	starterCount := queryAnswerExternalTestInt64(t, "SELECT COUNT(*) FROM owned_ships WHERE owner_id = $1 AND ship_id = $2", int64(response.GetUserId()), int64(201211))
	if starterCount != 1 {
		t.Fatalf("expected starter ship to exist")
	}
	belfastSecretaryCount := queryAnswerExternalTestInt64(t, "SELECT COUNT(*) FROM owned_ships WHERE owner_id = $1 AND ship_id = $2 AND is_secretary = TRUE", int64(response.GetUserId()), int64(202124))
	if belfastSecretaryCount != 1 {
		t.Fatalf("expected Belfast to be secretary")
	}
}

func TestCreateNewPlayerNameTooShort(t *testing.T) {
	loadCreatePlayerConfig(t, false, nil, "")
	client := &connection.Client{AuthArg2: 900002}
	payload := &protobuf.CS_10024{
		NickName: proto.String("abc"),
		ShipId:   proto.Uint32(201211),
		DeviceId: proto.String("device-101"),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.CreateNewPlayer(&buf, client); err != nil {
		t.Fatalf("CreateNewPlayer failed: %v", err)
	}
	response := &protobuf.SC_10025{}
	decodeResponsePacket(t, client, 10025, response)
	if response.GetResult() != 2012 {
		t.Fatalf("expected result 2012, got %d", response.GetResult())
	}
}

func TestCreateNewPlayerBlacklist(t *testing.T) {
	loadCreatePlayerConfig(t, false, []string{"mol"}, "")
	client := &connection.Client{AuthArg2: 900003}
	payload := &protobuf.CS_10024{
		NickName: proto.String("Molly"),
		ShipId:   proto.Uint32(201211),
		DeviceId: proto.String("device-102"),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.CreateNewPlayer(&buf, client); err != nil {
		t.Fatalf("CreateNewPlayer failed: %v", err)
	}
	response := &protobuf.SC_10025{}
	decodeResponsePacket(t, client, 10025, response)
	if response.GetResult() != 2013 {
		t.Fatalf("expected result 2013, got %d", response.GetResult())
	}
}

func TestCreateNewPlayerIllegalPattern(t *testing.T) {
	loadCreatePlayerConfig(t, false, nil, `[^a-zA-Z0-9]`)
	client := &connection.Client{AuthArg2: 900004}
	payload := &protobuf.CS_10024{
		NickName: proto.String("Molly!"),
		ShipId:   proto.Uint32(201211),
		DeviceId: proto.String("device-103"),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.CreateNewPlayer(&buf, client); err != nil {
		t.Fatalf("CreateNewPlayer failed: %v", err)
	}
	response := &protobuf.SC_10025{}
	decodeResponsePacket(t, client, 10025, response)
	if response.GetResult() != 2014 {
		t.Fatalf("expected result 2014, got %d", response.GetResult())
	}
}

func TestCreateNewPlayerDuplicateName(t *testing.T) {
	loadCreatePlayerConfig(t, false, nil, "")
	if err := orm.CreateCommanderRoot(910000, 910000, "Molly", 0, 0); err != nil {
		t.Fatalf("failed to seed commander: %v", err)
	}
	client := &connection.Client{AuthArg2: 900005}
	payload := &protobuf.CS_10024{
		NickName: proto.String("Molly"),
		ShipId:   proto.Uint32(201211),
		DeviceId: proto.String("device-104"),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.CreateNewPlayer(&buf, client); err != nil {
		t.Fatalf("CreateNewPlayer failed: %v", err)
	}
	response := &protobuf.SC_10025{}
	decodeResponsePacket(t, client, 10025, response)
	if response.GetResult() != 2015 {
		t.Fatalf("expected result 2015, got %d", response.GetResult())
	}
}

func TestCreateNewPlayerDuplicateDevice(t *testing.T) {
	loadCreatePlayerConfig(t, false, nil, "")
	if err := orm.UpsertDeviceAuthMap("device-105", 900006, 910001); err != nil {
		t.Fatalf("failed to seed device mapping: %v", err)
	}
	client := &connection.Client{AuthArg2: 900006}
	payload := &protobuf.CS_10024{
		NickName: proto.String("Molly"),
		ShipId:   proto.Uint32(201211),
		DeviceId: proto.String("device-105"),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.CreateNewPlayer(&buf, client); err != nil {
		t.Fatalf("CreateNewPlayer failed: %v", err)
	}
	response := &protobuf.SC_10025{}
	decodeResponsePacket(t, client, 10025, response)
	if response.GetResult() != 1011 {
		t.Fatalf("expected result 1011, got %d", response.GetResult())
	}
}

func TestJoinServerResolvesDeviceMapping(t *testing.T) {
	loadCreatePlayerConfig(t, false, nil, "")
	if err := orm.CreateCommanderRoot(920001, 920001, "Device Commander", 0, 0); err != nil {
		t.Fatalf("failed to seed commander: %v", err)
	}
	if err := orm.UpsertDeviceAuthMap("device-200", 900010, 920001); err != nil {
		t.Fatalf("failed to seed device mapping: %v", err)
	}
	client := &connection.Client{}
	payload := &protobuf.CS_10022{
		AccountId:    proto.Uint32(0),
		ServerTicket: proto.String(serverTicketPrefix),
		Platform:     proto.String("0"),
		Serverid:     proto.Uint32(1),
		CheckKey:     proto.String("check"),
		DeviceId:     proto.String("device-200"),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.JoinServer(&buf, client); err != nil {
		t.Fatalf("JoinServer failed: %v", err)
	}
	response := &protobuf.SC_10023{}
	decodeResponsePacket(t, client, 10023, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if response.GetUserId() != 920001 {
		t.Fatalf("expected user id 920001, got %d", response.GetUserId())
	}
}

func TestJoinServerResolvesServerTicket(t *testing.T) {
	loadCreatePlayerConfig(t, false, nil, "")
	if err := orm.CreateCommanderRoot(920002, 920002, "Ticket Commander", 0, 0); err != nil {
		t.Fatalf("failed to seed commander: %v", err)
	}
	if err := orm.CreateYostarusMap(900011, 920002); err != nil {
		t.Fatalf("failed to seed yostarus map: %v", err)
	}
	client := &connection.Client{}
	payload := &protobuf.CS_10022{
		AccountId:    proto.Uint32(0),
		ServerTicket: proto.String(fmt.Sprintf("%s:%d", serverTicketPrefix, 900011)),
		Platform:     proto.String("0"),
		Serverid:     proto.Uint32(1),
		CheckKey:     proto.String("check"),
		DeviceId:     proto.String(""),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.JoinServer(&buf, client); err != nil {
		t.Fatalf("JoinServer failed: %v", err)
	}
	response := &protobuf.SC_10023{}
	decodeResponsePacket(t, client, 10023, response)
	if response.GetUserId() != 920002 {
		t.Fatalf("expected user id 920002, got %d", response.GetUserId())
	}
}

func TestJoinServerSkipOnboarding(t *testing.T) {
	loadCreatePlayerConfig(t, true, nil, "")
	client := &connection.Client{}
	payload := &protobuf.CS_10022{
		AccountId:    proto.Uint32(0),
		ServerTicket: proto.String(fmt.Sprintf("%s:%d", serverTicketPrefix, 900030)),
		Platform:     proto.String("0"),
		Serverid:     proto.Uint32(1),
		CheckKey:     proto.String("check"),
		DeviceId:     proto.String(""),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.JoinServer(&buf, client); err != nil {
		t.Fatalf("JoinServer failed: %v", err)
	}
	response := &protobuf.SC_10023{}
	decodeResponsePacket(t, client, 10023, response)
	if response.GetUserId() == 0 {
		t.Fatalf("expected user id to be created")
	}
	if _, err := orm.GetYostarusMapByArg2(900030); err != nil {
		t.Fatalf("failed to fetch yostarus map: %v", err)
	}
}

func TestAuthConfirmSkipOnboarding(t *testing.T) {
	loadCreatePlayerConfig(t, true, nil, "")
	client := &connection.Client{}
	payload := &protobuf.CS_10020{
		LoginType: proto.Uint32(1),
		Arg1:      proto.String("yostarus"),
		Arg2:      proto.String("900020"),
		CheckKey:  proto.String("check"),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.Forge_SC10021(&buf, client); err != nil {
		t.Fatalf("Forge_SC10021 failed: %v", err)
	}
	response := &protobuf.SC_10021{}
	decodeResponsePacket(t, client, 10021, response)
	if response.GetAccountId() == 0 {
		t.Fatalf("expected account id to be created")
	}
	mapping, err := orm.GetYostarusMapByArg2(900020)
	if err != nil {
		t.Fatalf("failed to fetch yostarus map: %v", err)
	}
	if mapping.AccountID != response.GetAccountId() {
		t.Fatalf("expected mapping account %d, got %d", response.GetAccountId(), mapping.AccountID)
	}
}
