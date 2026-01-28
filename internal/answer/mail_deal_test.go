package answer_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

type clientCapture struct {
	packetId int
	payload  *protobuf.SC_30007
}

type collectionCapture struct {
	packetId int
	payload  *protobuf.SC_30005
}

func newTestClient(t *testing.T) *connection.Client {
	commanderID := uint32(time.Now().UnixNano())
	commander := orm.Commander{
		CommanderID: commanderID,
		AccountID:   commanderID,
		Name:        fmt.Sprintf("Mail Commander %d", commanderID),
	}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	if err := commander.Load(); err != nil {
		t.Fatalf("failed to load commander: %v", err)
	}
	return &connection.Client{Commander: &commander}
}

func insertMail(t *testing.T, commander *orm.Commander, title string, attachments []orm.MailAttachment, archived bool) *orm.Mail {
	mail := orm.Mail{
		ReceiverID: commander.CommanderID,
		Title:      title,
		Body:       "body",
		Read:       false,
		IsArchived: archived,
	}
	if err := orm.GormDB.Create(&mail).Error; err != nil {
		t.Fatalf("failed to create mail: %v", err)
	}
	for i := range attachments {
		attachments[i].MailID = mail.ID
		if err := orm.GormDB.Create(&attachments[i]).Error; err != nil {
			t.Fatalf("failed to create mail attachment: %v", err)
		}
	}
	if err := commander.Load(); err != nil {
		t.Fatalf("failed to reload commander: %v", err)
	}
	return commander.MailsMap[mail.ID]
}

func decodePacket(t *testing.T, client *connection.Client, expectedId int, message proto.Message) int {
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

func sendMailDeal(t *testing.T, client *connection.Client, payload *protobuf.CS_30006) clientCapture {
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.HandleMailDealCmd(&buf, client); err != nil {
		t.Fatalf("HandleMailDealCmd failed: %v", err)
	}
	response := &protobuf.SC_30007{}
	packetId := decodePacket(t, client, 30007, response)
	return clientCapture{packetId: packetId, payload: response}
}

func sendCollectionList(t *testing.T, client *connection.Client, payload *protobuf.CS_30004) collectionCapture {
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	client.Buffer.Reset()
	if _, _, err := answer.GetCollectionMailList(&buf, client); err != nil {
		t.Fatalf("GetCollectionMailList failed: %v", err)
	}
	response := &protobuf.SC_30005{}
	packetId := decodePacket(t, client, 30005, response)
	return collectionCapture{packetId: packetId, payload: response}
}

func TestMailDealAttachmentCollect(t *testing.T) {
	client := newTestClient(t)
	mail1 := insertMail(t, client.Commander, "mail-1", []orm.MailAttachment{
		{Type: 1, ItemID: 1, Quantity: 50},
		{Type: 2, ItemID: 45, Quantity: 2},
	}, false)
	mail2 := insertMail(t, client.Commander, "mail-2", []orm.MailAttachment{
		{Type: 2, ItemID: 45, Quantity: 3},
	}, false)
	matchList := []*protobuf.MATCH_EXPRESSION{
		{Type: proto.Uint32(1), ArgList: []uint32{mail1.ID, mail2.ID}},
	}
	payload := &protobuf.CS_30006{
		Cmd:       proto.Uint32(consts.MAIL_DEAL_CMDS_ATTACHMENT),
		MatchList: matchList,
	}
	response := sendMailDeal(t, client, payload)
	if response.packetId != 30007 {
		t.Fatalf("expected packet 30007, got %d", response.packetId)
	}
	if response.payload == nil {
		t.Fatal("expected payload")
	}
	if len(response.payload.MailIdList) != 2 {
		t.Fatalf("expected 2 mail ids, got %d", len(response.payload.MailIdList))
	}
	if len(response.payload.DropList) != 2 {
		t.Fatalf("expected 2 merged drops, got %d", len(response.payload.DropList))
	}
	var resourceDrop *protobuf.DROPINFO
	var itemDrop *protobuf.DROPINFO
	for _, drop := range response.payload.DropList {
		switch {
		case drop.GetType() == 1 && drop.GetId() == 1:
			resourceDrop = drop
		case drop.GetType() == 2 && drop.GetId() == 45:
			itemDrop = drop
		}
	}
	if resourceDrop == nil || resourceDrop.GetNumber() != 50 {
		t.Fatalf("expected gold drop of 50")
	}
	if itemDrop == nil || itemDrop.GetNumber() != 5 {
		t.Fatalf("expected item drop of 5")
	}
	if err := client.Commander.Load(); err != nil {
		t.Fatalf("failed to reload commander: %v", err)
	}
	if !client.Commander.MailsMap[mail1.ID].AttachmentsCollected || !client.Commander.MailsMap[mail2.ID].AttachmentsCollected {
		t.Fatalf("expected attachments collected")
	}
}

func TestMailDealOverflowPreview(t *testing.T) {
	client := newTestClient(t)
	mail := insertMail(t, client.Commander, "mail-3", []orm.MailAttachment{{Type: 1, ItemID: 1, Quantity: 20}}, false)
	payload := &protobuf.CS_30006{
		Cmd: proto.Uint32(consts.MAIL_DEAL_CMDS_OVERFLOW),
		MatchList: []*protobuf.MATCH_EXPRESSION{
			{Type: proto.Uint32(1), ArgList: []uint32{mail.ID}},
		},
	}
	response := sendMailDeal(t, client, payload)
	if response.packetId != 30007 {
		t.Fatalf("expected packet 30007, got %d", response.packetId)
	}
	if len(response.payload.DropList) != 1 {
		t.Fatalf("expected 1 drop, got %d", len(response.payload.DropList))
	}
	if mail.AttachmentsCollected {
		t.Fatalf("expected attachments to remain uncollected")
	}
}

func TestCollectionMailListPagination(t *testing.T) {
	client := newTestClient(t)
	archive1 := insertMail(t, client.Commander, "archive-1", []orm.MailAttachment{}, true)
	archive2 := insertMail(t, client.Commander, "archive-2", []orm.MailAttachment{}, true)
	insertMail(t, client.Commander, "inbox", []orm.MailAttachment{}, false)
	payload := &protobuf.CS_30004{
		IndexBegin: proto.Uint32(1),
		IndexEnd:   proto.Uint32(1),
	}
	response := sendCollectionList(t, client, payload)
	if response.packetId != 30005 {
		t.Fatalf("expected packet 30005, got %d", response.packetId)
	}
	if response.payload == nil {
		t.Fatal("expected payload")
	}
	if len(response.payload.MailList) != 1 {
		t.Fatalf("expected 1 mail, got %d", len(response.payload.MailList))
	}
	returnedId := response.payload.MailList[0].GetId()
	if returnedId != archive1.ID && returnedId != archive2.ID {
		t.Fatalf("expected archived mail, got %d", returnedId)
	}
}
