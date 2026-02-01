package answer

import (
	"os"
	"testing"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func setupRegisterTest(t *testing.T) *connection.Client {
	t.Helper()
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearTable(t, &orm.LocalAccount{})
	clearTable(t, &orm.YostarusMap{})
	return &connection.Client{}
}

func decodeRegisterResponse(t *testing.T, client *connection.Client, expectedID int, message proto.Message) {
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

func TestRegisterAccountSuccess(t *testing.T) {
	client := setupRegisterTest(t)
	payload := &protobuf.CS_10001{
		Account:  proto.String("testuser"),
		Password: proto.String("pass"),
		MailBox:  proto.String("mail"),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := RegisterAccount(&buf, client); err != nil {
		t.Fatalf("RegisterAccount failed: %v", err)
	}
	response := &protobuf.SC_10002{}
	decodeRegisterResponse(t, client, 10002, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	var stored orm.LocalAccount
	if err := orm.GormDB.Where("account = ?", "testuser").First(&stored).Error; err != nil {
		t.Fatalf("fetch local account: %v", err)
	}
	if stored.Arg2 == 0 {
		t.Fatalf("expected non-zero arg2")
	}
	if stored.MailBox != "mail" {
		t.Fatalf("expected mailbox mail, got %s", stored.MailBox)
	}
}

func TestRegisterAccountDuplicate(t *testing.T) {
	client := setupRegisterTest(t)
	if err := orm.GormDB.Create(&orm.LocalAccount{Arg2: 900010, Account: "testuser", Password: "pass", MailBox: ""}).Error; err != nil {
		t.Fatalf("seed local account: %v", err)
	}
	payload := &protobuf.CS_10001{
		Account:  proto.String("testuser"),
		Password: proto.String("pass"),
		MailBox:  proto.String(""),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := RegisterAccount(&buf, client); err != nil {
		t.Fatalf("RegisterAccount failed: %v", err)
	}
	response := &protobuf.SC_10002{}
	decodeRegisterResponse(t, client, 10002, response)
	if response.GetResult() != 1011 {
		t.Fatalf("expected result 1011, got %d", response.GetResult())
	}
}

func TestRegisterAccountNumericOnly(t *testing.T) {
	client := setupRegisterTest(t)
	payload := &protobuf.CS_10001{
		Account:  proto.String("12345"),
		Password: proto.String("pass"),
		MailBox:  proto.String(""),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := RegisterAccount(&buf, client); err != nil {
		t.Fatalf("RegisterAccount failed: %v", err)
	}
	response := &protobuf.SC_10002{}
	decodeRegisterResponse(t, client, 10002, response)
	if response.GetResult() != 1012 {
		t.Fatalf("expected result 1012, got %d", response.GetResult())
	}
}

func TestRegisterAccountEmpty(t *testing.T) {
	client := setupRegisterTest(t)
	payload := &protobuf.CS_10001{
		Account:  proto.String(""),
		Password: proto.String("pass"),
		MailBox:  proto.String(""),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := RegisterAccount(&buf, client); err != nil {
		t.Fatalf("RegisterAccount failed: %v", err)
	}
	response := &protobuf.SC_10002{}
	decodeRegisterResponse(t, client, 10002, response)
	if response.GetResult() != 1010 {
		t.Fatalf("expected result 1010, got %d", response.GetResult())
	}
}
