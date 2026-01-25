package answer

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

var instagramChatCommanderID uint32 = 7000

func setupInstagramChatTest(t *testing.T) (*connection.Client, uint32) {
	t.Helper()
	initJuustagramHandlerTestDB(t)
	clearTable(t, &orm.JuustagramReply{})
	clearTable(t, &orm.JuustagramChatGroup{})
	clearTable(t, &orm.JuustagramGroup{})
	clearTable(t, &orm.ConfigEntry{})
	clearTable(t, &orm.OwnedResource{})
	clearTable(t, &orm.Resource{})
	clearTable(t, &orm.Commander{})
	commanderID := atomic.AddUint32(&instagramChatCommanderID, 1)
	commander := orm.Commander{
		CommanderID: commanderID,
		AccountID:   1,
		Level:       1,
		Exp:         0,
		Name:        "Juustagram Tester",
		LastLogin:   time.Now().UTC(),
	}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}
	resource := orm.Resource{ID: 1, ItemID: 0, Name: "Coins"}
	if err := orm.GormDB.Create(&resource).Error; err != nil {
		t.Fatalf("create resource: %v", err)
	}
	client := &connection.Client{Commander: &orm.Commander{CommanderID: commanderID}}
	return client, commanderID
}

func TestInstagramChatReplyCreatesEntry(t *testing.T) {
	client, commanderID := setupInstagramChatTest(t)
	if _, err := orm.CreateJuustagramGroup(commanderID, 960007, 1); err != nil {
		t.Fatalf("create juustagram group: %v", err)
	}
	payload := protobuf.CS_11712{
		ChatGroupId: proto.Uint32(1),
		ChatId:      proto.Uint32(1),
		Value:       proto.Uint32(2),
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	_, packetID, err := InstagramChatReply(&buffer, client)
	if err != nil {
		t.Fatalf("instagram chat reply failed: %v", err)
	}
	if packetID != 11713 {
		t.Fatalf("expected packet 11713, got %d", packetID)
	}
	var response protobuf.SC_11713
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	group, err := orm.GetJuustagramChatGroup(commanderID, 1)
	if err != nil {
		t.Fatalf("load juustagram chat group: %v", err)
	}
	if len(group.ReplyList) != 1 {
		t.Fatalf("expected 1 reply, got %d", len(group.ReplyList))
	}
	if group.ReplyList[0].Key != 1 || group.ReplyList[0].Value != 2 {
		t.Fatalf("unexpected reply values")
	}
}

func TestInstagramChatReplyRedPacketRewards(t *testing.T) {
	client, commanderID := setupInstagramChatTest(t)
	if _, err := orm.CreateJuustagramGroup(commanderID, 960007, 1); err != nil {
		t.Fatalf("create juustagram group: %v", err)
	}
	seedConfigEntry(t, "ShareCfg/activity_ins_redpackage.json", "1000", `{"id":1000,"type":1,"content":[1,1,5]}`)
	payload := protobuf.CS_11712{
		ChatGroupId: proto.Uint32(1),
		ChatId:      proto.Uint32(1),
		Value:       proto.Uint32(1000),
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := InstagramChatReply(&buffer, client); err != nil {
		t.Fatalf("instagram chat reply failed: %v", err)
	}
	var response protobuf.SC_11713
	decodeResponse(t, client, &response)
	if len(response.GetDropList()) != 1 {
		t.Fatalf("expected 1 drop, got %d", len(response.GetDropList()))
	}
	if response.GetDropList()[0].GetType() != 1 || response.GetDropList()[0].GetId() != 1 || response.GetDropList()[0].GetNumber() != 5 {
		t.Fatalf("unexpected drop contents")
	}
	var resource orm.OwnedResource
	if err := orm.GormDB.First(&resource, "commander_id = ? AND resource_id = ?", commanderID, 1).Error; err != nil {
		t.Fatalf("load resource: %v", err)
	}
	if resource.Amount != 5 {
		t.Fatalf("expected resource amount 5, got %d", resource.Amount)
	}
}

func TestInstagramChatSetSkin(t *testing.T) {
	client, commanderID := setupInstagramChatTest(t)
	if _, err := orm.CreateJuustagramGroup(commanderID, 960007, 1); err != nil {
		t.Fatalf("create juustagram group: %v", err)
	}
	payload := protobuf.CS_11714{
		GroupId: proto.Uint32(960007),
		SkinId:  proto.Uint32(12345),
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := InstagramChatSetSkin(&buffer, client); err != nil {
		t.Fatalf("instagram chat set skin failed: %v", err)
	}
	group, err := orm.GetJuustagramGroup(commanderID, 960007)
	if err != nil {
		t.Fatalf("load juustagram group: %v", err)
	}
	if group.SkinID != 12345 {
		t.Fatalf("expected skin id 12345, got %d", group.SkinID)
	}
}

func TestInstagramChatSetCare(t *testing.T) {
	client, commanderID := setupInstagramChatTest(t)
	if _, err := orm.CreateJuustagramGroup(commanderID, 960007, 1); err != nil {
		t.Fatalf("create juustagram group: %v", err)
	}
	payload := protobuf.CS_11716{
		GroupId: proto.Uint32(960007),
		Value:   proto.Uint32(1),
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := InstagramChatSetCare(&buffer, client); err != nil {
		t.Fatalf("instagram chat set care failed: %v", err)
	}
	group, err := orm.GetJuustagramGroup(commanderID, 960007)
	if err != nil {
		t.Fatalf("load juustagram group: %v", err)
	}
	if group.Favorite != 1 {
		t.Fatalf("expected favorite 1, got %d", group.Favorite)
	}
}

func TestInstagramChatSetTopic(t *testing.T) {
	client, commanderID := setupInstagramChatTest(t)
	if _, err := orm.CreateJuustagramGroup(commanderID, 960007, 1); err != nil {
		t.Fatalf("create juustagram group: %v", err)
	}
	if _, err := orm.CreateJuustagramChatGroup(commanderID, 960007, 2, orm.DefaultJuustagramOpTime()); err != nil {
		t.Fatalf("create juustagram chat group: %v", err)
	}
	payload := protobuf.CS_11718{
		ChatGroupId: proto.Uint32(2),
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := InstagramChatSetTopic(&buffer, client); err != nil {
		t.Fatalf("instagram chat set topic failed: %v", err)
	}
	group, err := orm.GetJuustagramGroup(commanderID, 960007)
	if err != nil {
		t.Fatalf("load juustagram group: %v", err)
	}
	if group.CurChatGroup != 2 {
		t.Fatalf("expected cur_chat_group 2, got %d", group.CurChatGroup)
	}
}

func TestInstagramChatActivateTopic(t *testing.T) {
	client, commanderID := setupInstagramChatTest(t)
	seedConfigEntry(t, "ShareCfg/activity_ins_chat_group.json", "1", `{"id":1,"ship_group":960007}`)
	seedConfigEntry(t, "ShareCfg/activity_ins_chat_group.json", "2", `{"id":2,"ship_group":960007}`)
	payload := protobuf.CS_11722{
		ChatGroupIdList: []uint32{1, 2},
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := InstagramChatActivateTopic(&buffer, client); err != nil {
		t.Fatalf("instagram chat activate topic failed: %v", err)
	}
	var response protobuf.SC_11723
	decodeResponse(t, client, &response)
	if len(response.GetResultList()) != 2 {
		t.Fatalf("expected 2 results, got %d", len(response.GetResultList()))
	}
	for _, result := range response.GetResultList() {
		if result != 0 {
			t.Fatalf("expected result 0, got %d", result)
		}
	}
	if _, err := orm.GetJuustagramChatGroup(commanderID, 1); err != nil {
		t.Fatalf("load chat group 1: %v", err)
	}
	if _, err := orm.GetJuustagramChatGroup(commanderID, 2); err != nil {
		t.Fatalf("load chat group 2: %v", err)
	}
}
