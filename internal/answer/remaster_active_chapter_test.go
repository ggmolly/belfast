package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestRemasterSetActiveChapterStoresActiveID(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.RemasterState{})

	payload := protobuf.CS_13501{ActiveId: proto.Uint32(42)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := RemasterSetActiveChapter(&buffer, client); err != nil {
		t.Fatalf("set active chapter failed: %v", err)
	}

	var response protobuf.SC_13502
	decodeResponse(t, client, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected success result")
	}

	var state orm.RemasterState
	if err := orm.GormDB.First(&state, "commander_id = ?", client.Commander.CommanderID).Error; err != nil {
		t.Fatalf("load remaster state: %v", err)
	}
	if state.ActiveChapterID != 42 {
		t.Fatalf("expected active chapter id 42, got %d", state.ActiveChapterID)
	}
}
