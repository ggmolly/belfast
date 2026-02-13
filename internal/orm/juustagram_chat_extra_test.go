package orm

import (
	"errors"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
)

func TestJuustagramGroupAndChatFlows(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &JuustagramReply{})
	clearTable(t, &JuustagramChatGroup{})
	clearTable(t, &JuustagramGroup{})

	group, err := CreateJuustagramGroup(1, 10, 100)
	if err != nil {
		t.Fatalf("create juustagram group: %v", err)
	}
	if group.GroupID != 10 || len(group.ChatGroups) != 1 {
		t.Fatalf("unexpected group data")
	}
	groups, err := GetJuustagramGroups(1)
	if err != nil || len(groups) != 1 {
		t.Fatalf("get juustagram groups: %v", err)
	}
	list, total, err := ListJuustagramGroups(1, 0, 10)
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatalf("list juustagram groups: %v", err)
	}

	if err := UpdateJuustagramGroup(1, 10, nil, nil, nil); err != nil {
		t.Fatalf("update no-op: %v", err)
	}
	skin := uint32(5)
	if err := UpdateJuustagramGroup(1, 10, &skin, nil, nil); err != nil {
		t.Fatalf("update group: %v", err)
	}
	if err := UpdateJuustagramGroup(1, 99, &skin, nil, nil); !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("expected not found update error")
	}

	chat, err := CreateJuustagramChatGroup(1, 10, 101, 55)
	if err != nil {
		t.Fatalf("create chat group: %v", err)
	}
	if chat.ChatGroupID != 101 {
		t.Fatalf("unexpected chat group")
	}

	if _, err := AddJuustagramChatReply(1, 101, 7, 9, 100); err != nil {
		t.Fatalf("add chat reply: %v", err)
	}
	chatLoaded, err := GetJuustagramChatGroup(1, 101)
	if err != nil || len(chatLoaded.ReplyList) != 1 {
		t.Fatalf("get chat group: %v", err)
	}

	if err := SetJuustagramChatGroupRead(1, []uint32{101}); err != nil {
		t.Fatalf("set chat read: %v", err)
	}
	if err := MarkJuustagramChatGroupsRead(1, nil); err != nil {
		t.Fatalf("mark chat groups read: %v", err)
	}
	if err := SetJuustagramCurrentChatGroup(1, 101); err != nil {
		t.Fatalf("set current chat group: %v", err)
	}

	ensured, err := EnsureJuustagramGroupExists(1, 11, 200)
	if err != nil || ensured.GroupID != 11 {
		t.Fatalf("ensure group exists: %v", err)
	}
}

func TestJuustagramValidationHelpers(t *testing.T) {
	if err := ValidateJuustagramChatGroupIDs(nil); err == nil {
		t.Fatalf("expected error for empty list")
	}
	if err := ValidateJuustagramChatGroupIDs([]uint32{0}); err == nil {
		t.Fatalf("expected error for zero id")
	}
	if err := ValidateJuustagramChatGroupIDs([]uint32{1, 2}); err != nil {
		t.Fatalf("expected valid ids")
	}
	if DefaultJuustagramOpTime() == 0 {
		t.Fatalf("expected default op time")
	}
}
