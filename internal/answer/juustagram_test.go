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
	if err := orm.CreateCommanderRoot(commanderID, commanderID, fmt.Sprintf("Juustagram Tester %d", commanderID), 0, 0); err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	commander := &orm.Commander{CommanderID: commanderID}
	if err := commander.Load(); err != nil {
		t.Fatalf("failed to load commander: %v", err)
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
	createdGroup, err := orm.CreateJuustagramGroup(commanderID, groupID, chatGroupIDs[0])
	if err != nil {
		t.Fatalf("failed to create juustagram group: %v", err)
	}
	group = *createdGroup
	for i, chatGroupID := range chatGroupIDs {
		if i == 0 {
			continue
		}
		if _, err := orm.CreateJuustagramChatGroup(commanderID, groupID, chatGroupID, 0); err != nil {
			t.Fatalf("failed to create juustagram chat group: %v", err)
		}
	}
	return group
}

func TestJuustagramDataLoadsStoredGroups(t *testing.T) {
	commander := createTestCommander(t, 9001)
	group := createJuustagramGroup(t, commander.CommanderID, 960007, []uint32{1})
	chatGroup, err := orm.GetJuustagramChatGroup(commander.CommanderID, 1)
	if err != nil {
		t.Fatalf("failed to load juustagram chat group: %v", err)
	}
	if _, err := orm.AddJuustagramChatReply(commander.CommanderID, chatGroup.ChatGroupID, 33, 44, 0); err != nil {
		t.Fatalf("failed to create first reply: %v", err)
	}
	if _, err := orm.AddJuustagramChatReply(commander.CommanderID, chatGroup.ChatGroupID, 11, 22, 0); err != nil {
		t.Fatalf("failed to create second reply: %v", err)
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
	updated, err := orm.GetJuustagramChatGroup(commander.CommanderID, 2)
	if err != nil {
		t.Fatalf("failed to load updated chat group: %v", err)
	}
	if updated.ReadFlag != 1 {
		t.Fatalf("expected read flag 1, got %d", updated.ReadFlag)
	}
	untouched, err := orm.GetJuustagramChatGroup(commander.CommanderID, 1)
	if err != nil {
		t.Fatalf("failed to load untouched chat group: %v", err)
	}
	if untouched.ReadFlag != 0 {
		t.Fatalf("expected read flag 0, got %d", untouched.ReadFlag)
	}
}
