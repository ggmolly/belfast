package answer_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

type exchangeRewardPayload struct {
	Type  uint32 `json:"type"`
	ID    uint32 `json:"id"`
	Count uint32 `json:"count"`
}

var exchangeCommanderID uint32 = 200

func resetExchangeTables(t *testing.T) {
	if err := orm.GormDB.Exec("DELETE FROM exchange_code_redeems").Error; err != nil {
		t.Fatalf("failed to clear exchange_code_redeems: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM exchange_codes").Error; err != nil {
		t.Fatalf("failed to clear exchange_codes: %v", err)
	}
}

func newExchangeCommander(t *testing.T) *orm.Commander {
	commander := &orm.Commander{
		CommanderID: exchangeCommanderID,
		AccountID:   exchangeCommanderID,
		Name:        fmt.Sprintf("Exchange Tester %d", exchangeCommanderID),
	}
	exchangeCommanderID++

	if err := orm.GormDB.Create(commander).Error; err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}

	commander.OwnedResourcesMap = make(map[uint32]*orm.OwnedResource)
	commander.CommanderItemsMap = make(map[uint32]*orm.CommanderItem)
	commander.OwnedShipsMap = make(map[uint32]*orm.OwnedShip)
	commander.OwnedSkinsMap = make(map[uint32]*orm.OwnedSkin)

	resource := orm.OwnedResource{
		CommanderID: commander.CommanderID,
		ResourceID:  1,
		Amount:      100,
	}
	if err := orm.GormDB.Create(&resource).Error; err != nil {
		t.Fatalf("failed to create resource: %v", err)
	}
	commander.OwnedResources = append(commander.OwnedResources, resource)
	commander.OwnedResourcesMap[1] = &commander.OwnedResources[0]

	return commander
}

func createExchangeCode(t *testing.T, code string, quota int, rewards []exchangeRewardPayload) orm.ExchangeCode {
	payload, err := json.Marshal(rewards)
	if err != nil {
		t.Fatalf("failed to marshal rewards: %v", err)
	}
	exchangeCode := orm.ExchangeCode{
		Code:    code,
		Quota:   quota,
		Rewards: payload,
	}
	if err := orm.GormDB.Create(&exchangeCode).Error; err != nil {
		t.Fatalf("failed to create exchange code: %v", err)
	}
	return exchangeCode
}

func TestExchangeCodeRedeemSuccessAndRepeat(t *testing.T) {
	resetExchangeTables(t)
	commander := newExchangeCommander(t)
	exchangeCode := createExchangeCode(t, "PROMO123", 1, []exchangeRewardPayload{
		{Type: consts.DROP_TYPE_RESOURCE, ID: 1, Count: 10},
	})
	client := &connection.Client{Commander: commander}
	payload := &protobuf.CS_11508{Key: proto.String("promo123"), Platform: proto.String("")}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	n, packetId, err := answer.ExchangeCodeRedeem(&buf, client)
	if err != nil {
		t.Fatalf("ExchangeCodeRedeem failed: %v", err)
	}
	if client.Buffer.Len() == 0 {
		t.Fatalf("expected response buffer, got empty (n=%d packet=%d)", n, packetId)
	}
	if len(commander.Mails) == 0 {
		t.Fatalf("expected mail to be created")
	}
	buffer := client.Buffer.Bytes()
	packetId = packets.GetPacketId(0, &buffer)
	if packetId != 30001 {
		t.Fatalf("expected packet 30001, got %d", packetId)
	}
	packetSize := packets.GetPacketSize(0, &buffer) + 2
	if len(buffer) < packetSize {
		t.Fatalf("expected packet size %d, got %d", packetSize, len(buffer))
	}
	payloadStart := packets.HEADER_SIZE
	payloadEnd := payloadStart + (packetSize - packets.HEADER_SIZE)
	mailbox := &protobuf.SC_30001{}
	if err := proto.Unmarshal(buffer[payloadStart:payloadEnd], mailbox); err != nil {
		t.Fatalf("failed to unmarshal mailbox response: %v", err)
	}
	if mailbox.GetUnreadNumber() != 1 || mailbox.GetTotalNumber() != 1 {
		t.Fatalf("expected mailbox counts 1/1, got %d/%d", mailbox.GetUnreadNumber(), mailbox.GetTotalNumber())
	}
	if len(buffer) <= packetSize {
		t.Fatalf("expected second packet")
	}
	offset := packetSize
	packetId = packets.GetPacketId(offset, &buffer)
	if packetId != 11509 {
		t.Fatalf("expected packet 11509, got %d", packetId)
	}
	packetSize = packets.GetPacketSize(offset, &buffer) + 2
	if len(buffer) < offset+packetSize {
		t.Fatalf("expected packet size %d, got %d", offset+packetSize, len(buffer))
	}
	payloadStart = offset + packets.HEADER_SIZE
	payloadEnd = payloadStart + (packetSize - packets.HEADER_SIZE)
	response := &protobuf.SC_11509{}
	if err := proto.Unmarshal(buffer[payloadStart:payloadEnd], response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	client.Buffer.Reset()
	if commander.OwnedResourcesMap[1].Amount != 100 {
		t.Fatalf("expected resource amount 100, got %d", commander.OwnedResourcesMap[1].Amount)
	}
	if len(commander.Mails) != 1 {
		t.Fatalf("expected 1 mail, got %d", len(commander.Mails))
	}
	mail := commander.Mails[0]
	if len(mail.Attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(mail.Attachments))
	}
	attachment := mail.Attachments[0]
	if attachment.Type != consts.DROP_TYPE_RESOURCE || attachment.ItemID != 1 || attachment.Quantity != 10 {
		t.Fatalf("unexpected attachment: type=%d id=%d quantity=%d", attachment.Type, attachment.ItemID, attachment.Quantity)
	}

	var updatedCode orm.ExchangeCode
	if err := orm.GormDB.First(&updatedCode, exchangeCode.ID).Error; err != nil {
		t.Fatalf("failed to load exchange code: %v", err)
	}
	if updatedCode.Quota != 0 {
		t.Fatalf("expected quota 0, got %d", updatedCode.Quota)
	}

	buf, err = proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.ExchangeCodeRedeem(&buf, client); err != nil {
		t.Fatalf("ExchangeCodeRedeem failed: %v", err)
	}
	response = &protobuf.SC_11509{}
	decodeTestPacket(t, client, 11509, response)
	if response.GetResult() != 1 {
		t.Fatalf("expected result 1, got %d", response.GetResult())
	}
}

func TestExchangeCodeRedeemInvalidCode(t *testing.T) {
	resetExchangeTables(t)
	commander := newExchangeCommander(t)
	client := &connection.Client{Commander: commander}
	payload := &protobuf.CS_11508{Key: proto.String("unknown"), Platform: proto.String("")}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.ExchangeCodeRedeem(&buf, client); err != nil {
		t.Fatalf("ExchangeCodeRedeem failed: %v", err)
	}
	response := &protobuf.SC_11509{}
	decodeTestPacket(t, client, 11509, response)
	if response.GetResult() != 1 {
		t.Fatalf("expected result 1, got %d", response.GetResult())
	}
}
