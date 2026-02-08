package answer

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func decodeGuildChatPacket(t *testing.T, client *connection.Client, expectedId int, message proto.Message) int {
	buffer := client.Buffer.Bytes()
	if len(buffer) == 0 {
		t.Fatalf("expected response buffer")
	}
	packetId := packets.GetPacketId(0, &buffer)
	if packetId != expectedId {
		t.Fatalf("expected packet %d, got %d", expectedId, packetId)
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
	return packetId
}

func setupGuildChatTest(t *testing.T) (*connection.Server, *connection.Client, *connection.Client) {
	t.Helper()
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearTable(t, &orm.GuildChatMessage{})
	clearTable(t, &orm.Commander{})

	server := connection.NewServer("127.0.0.1", 0, func(pkt *[]byte, c *connection.Client, size int) {})
	commanderID := uint32(time.Now().UnixNano() % 100000)
	commander := orm.Commander{
		CommanderID:         commanderID,
		AccountID:           commanderID,
		Name:                fmt.Sprintf("Guild Sender %d", commanderID),
		Level:               20,
		DisplayIconID:       1001,
		DisplaySkinID:       1001,
		SelectedIconFrameID: 200,
		SelectedChatFrameID: 300,
		DisplayIconThemeID:  400,
	}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}

	client := &connection.Client{Commander: &commander, Hash: 1}
	server.AddClient(client)

	otherCommander := orm.Commander{CommanderID: commanderID + 1, AccountID: commanderID + 1, Name: "Guild Listener", Level: 10}
	if err := orm.GormDB.Create(&otherCommander).Error; err != nil {
		t.Fatalf("create listener commander: %v", err)
	}
	listener := &connection.Client{Commander: &otherCommander, Hash: 2}
	server.AddClient(listener)

	return server, client, listener
}

func TestGuildSendMessageBroadcasts(t *testing.T) {
	_, client, listener := setupGuildChatTest(t)
	payload := protobuf.CS_60007{Chat: proto.String("hello guild")}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := GuildSendMessage(&buffer, client); err != nil {
		t.Fatalf("GuildSendMessage failed: %v", err)
	}

	var response protobuf.SC_60008
	decodeGuildChatPacket(t, client, 60008, &response)
	if response.GetChat().GetContent() != "hello guild" {
		t.Fatalf("expected chat content to match")
	}
	if response.GetChat().GetPlayer().GetId() != client.Commander.CommanderID {
		t.Fatalf("expected sender id to match")
	}

	var listenerResponse protobuf.SC_60008
	decodeGuildChatPacket(t, listener, 60008, &listenerResponse)
	if listenerResponse.GetChat().GetContent() != "hello guild" {
		t.Fatalf("expected listener to receive chat")
	}

	var count int64
	if err := orm.GormDB.Model(&orm.GuildChatMessage{}).Count(&count).Error; err != nil {
		t.Fatalf("count guild chat: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 chat message, got %d", count)
	}
}

func TestGuildSendMessageBroadcastsRegistrationPin(t *testing.T) {
	_, client, listener := setupGuildChatTest(t)
	payload := protobuf.CS_60007{Chat: proto.String("B-123456")}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := GuildSendMessage(&buffer, client); err != nil {
		t.Fatalf("GuildSendMessage failed: %v", err)
	}

	var response protobuf.SC_60008
	decodeGuildChatPacket(t, client, 60008, &response)
	if response.GetChat().GetContent() != "B-123456" {
		t.Fatalf("expected chat content to match")
	}

	var listenerResponse protobuf.SC_60008
	decodeGuildChatPacket(t, listener, 60008, &listenerResponse)
	if listenerResponse.GetChat().GetContent() != "B-123456" {
		t.Fatalf("expected listener to receive chat")
	}

	var count int64
	if err := orm.GormDB.Model(&orm.GuildChatMessage{}).Count(&count).Error; err != nil {
		t.Fatalf("count guild chat: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 chat message, got %d", count)
	}
}

func TestCommanderGuildChatHistory(t *testing.T) {
	_, client, _ := setupGuildChatTest(t)

	base := time.Date(2026, time.January, 1, 10, 0, 0, 0, time.UTC)
	if _, err := orm.CreateGuildChatMessage(guildChatPlaceholderID, client.Commander.CommanderID, "first", base); err != nil {
		t.Fatalf("create message 1: %v", err)
	}
	if _, err := orm.CreateGuildChatMessage(guildChatPlaceholderID, client.Commander.CommanderID, "second", base.Add(2*time.Minute)); err != nil {
		t.Fatalf("create message 2: %v", err)
	}

	payload := protobuf.CS_60100{Count: proto.Uint32(2)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := CommanderGuildChat(&buffer, client); err != nil {
		t.Fatalf("CommanderGuildChat failed: %v", err)
	}

	var response protobuf.SC_60101
	decodeGuildChatPacket(t, client, 60101, &response)
	if len(response.GetChatList()) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(response.GetChatList()))
	}
	if response.GetChatList()[0].GetContent() != "first" || response.GetChatList()[1].GetContent() != "second" {
		t.Fatalf("unexpected chat order")
	}
}
