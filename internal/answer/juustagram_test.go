package answer_test

import (
	"fmt"
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func createTestCommander(t *testing.T, commanderID uint32) *orm.Commander {
	commander := &orm.Commander{
		CommanderID: commanderID,
		AccountID:   commanderID,
		Name:        fmt.Sprintf("Juustagram Tester %d", commanderID),
	}
	if err := orm.GormDB.Create(commander).Error; err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	return commander
}

func createJuustagramGroup(t *testing.T, commanderID uint32, groupID uint32, chatGroupIDs []uint32) orm.JuustagramGroup {
	group := orm.JuustagramGroup{
		CommanderID:  commanderID,
		GroupID:      groupID,
		SkinID:       0,
		Favorite:     0,
		CurChatGroup: chatGroupIDs[0],
	}
	if err := orm.GormDB.Create(&group).Error; err != nil {
		t.Fatalf("failed to create juustagram group: %v", err)
	}
	for _, chatGroupID := range chatGroupIDs {
		chatGroup := orm.JuustagramChatGroup{
			CommanderID:   commanderID,
			GroupRecordID: group.ID,
			ChatGroupID:   chatGroupID,
			OpTime:        0,
			ReadFlag:      0,
		}
		if err := orm.GormDB.Create(&chatGroup).Error; err != nil {
			t.Fatalf("failed to create juustagram chat group: %v", err)
		}
	}
	return group
}

func TestJuustagramDataLoadsStoredGroups(t *testing.T) {
	commander := createTestCommander(t, 9001)
	group := createJuustagramGroup(t, commander.CommanderID, 960007, []uint32{1})
	chatGroup := orm.JuustagramChatGroup{}
	if err := orm.GormDB.Where("commander_id = ? AND chat_group_id = ?", commander.CommanderID, uint32(1)).First(&chatGroup).Error; err != nil {
		t.Fatalf("failed to load juustagram chat group: %v", err)
	}
	firstReply := orm.JuustagramReply{
		ChatGroupRecordID: chatGroup.ID,
		Sequence:          2,
		Key:               11,
		Value:             22,
	}
	secondReply := orm.JuustagramReply{
		ChatGroupRecordID: chatGroup.ID,
		Sequence:          1,
		Key:               33,
		Value:             44,
	}
	if err := orm.GormDB.Create(&firstReply).Error; err != nil {
		t.Fatalf("failed to create reply: %v", err)
	}
	if err := orm.GormDB.Create(&secondReply).Error; err != nil {
		t.Fatalf("failed to create reply: %v", err)
	}

	client := &connection.Client{Commander: commander}
	buffer := []byte{}
	if _, _, err := answer.JuustagramData(&buffer, client); err != nil {
		t.Fatalf("JuustagramData failed: %v", err)
	}
	response := &protobuf.SC_11711{}
	decodeTestPacket(t, client, 11711, response)
	if len(response.GetGroups()) != 1 {
		t.Fatalf("expected 1 group, got %d", len(response.GetGroups()))
	}
	loadedGroup := response.GetGroups()[0]
	if loadedGroup.GetId() != group.GroupID {
		t.Fatalf("expected group id %d, got %d", group.GroupID, loadedGroup.GetId())
	}
	if loadedGroup.GetCurChatGroup() != group.CurChatGroup {
		t.Fatalf("expected current chat group %d, got %d", group.CurChatGroup, loadedGroup.GetCurChatGroup())
	}
	if len(loadedGroup.GetChatGroupList()) != 1 {
		t.Fatalf("expected 1 chat group, got %d", len(loadedGroup.GetChatGroupList()))
	}
	loadedChat := loadedGroup.GetChatGroupList()[0]
	if loadedChat.GetId() != 1 {
		t.Fatalf("expected chat group id 1, got %d", loadedChat.GetId())
	}
	if len(loadedChat.GetReplyList()) != 2 {
		t.Fatalf("expected 2 replies, got %d", len(loadedChat.GetReplyList()))
	}
	if loadedChat.GetReplyList()[0].GetKey() != 33 || loadedChat.GetReplyList()[0].GetValue() != 44 {
		t.Fatalf("expected first reply to be sequence 1")
	}
	if loadedChat.GetReplyList()[1].GetKey() != 11 || loadedChat.GetReplyList()[1].GetValue() != 22 {
		t.Fatalf("expected second reply to be sequence 2")
	}
}

func TestJuustagramReadTipPersists(t *testing.T) {
	commander := createTestCommander(t, 9002)
	createJuustagramGroup(t, commander.CommanderID, 960007, []uint32{1, 2})
	client := &connection.Client{Commander: commander}
	payload := &protobuf.CS_11720{ChatGroupIdList: []uint32{2}}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.JuustagramReadTip(&buf, client); err != nil {
		t.Fatalf("JuustagramReadTip failed: %v", err)
	}
	var updated orm.JuustagramChatGroup
	if err := orm.GormDB.Where("commander_id = ? AND chat_group_id = ?", commander.CommanderID, uint32(2)).First(&updated).Error; err != nil {
		t.Fatalf("failed to load updated chat group: %v", err)
	}
	if updated.ReadFlag != 1 {
		t.Fatalf("expected read flag 1, got %d", updated.ReadFlag)
	}
	var untouched orm.JuustagramChatGroup
	if err := orm.GormDB.Where("commander_id = ? AND chat_group_id = ?", commander.CommanderID, uint32(1)).First(&untouched).Error; err != nil {
		t.Fatalf("failed to load untouched chat group: %v", err)
	}
	if untouched.ReadFlag != 0 {
		t.Fatalf("expected read flag 0, got %d", untouched.ReadFlag)
	}
}
