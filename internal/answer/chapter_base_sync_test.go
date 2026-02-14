package answer_test

import (
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestChapterBaseSyncNoState(t *testing.T) {
	if err := orm.CreateCommanderRoot(4242001, 4242001, "Chapter Base Sync", 0, 0); err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	commander, err := orm.GetCommanderCoreByID(4242001)
	if err != nil {
		t.Fatalf("failed to load commander: %v", err)
	}
	client := &connection.Client{Commander: commander}

	buf := []byte{}
	if _, _, err := answer.ChapterBaseSync(&buf, client); err != nil {
		t.Fatalf("ChapterBaseSync failed: %v", err)
	}

	response := &protobuf.SC_13000{}
	decodeTestPacket(t, client, 13000, response)
	if response.GetDailyRepairCount() != 0 {
		t.Fatalf("expected daily repair count 0, got %d", response.GetDailyRepairCount())
	}
	if response.GetCurrentChapter() != nil {
		t.Fatalf("expected no current chapter")
	}
}

func TestChapterBaseSyncWithState(t *testing.T) {
	if err := orm.CreateCommanderRoot(4242002, 4242002, "Chapter Base Sync 2", 0, 0); err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	commander, err := orm.GetCommanderCoreByID(4242002)
	if err != nil {
		t.Fatalf("failed to load commander: %v", err)
	}
	client := &connection.Client{Commander: commander}

	current := &protobuf.CURRENTCHAPTERINFO{
		Id:                  proto.Uint32(5001),
		Time:                proto.Uint32(0),
		Round:               proto.Uint32(1),
		ChapterHp:           proto.Uint32(100),
		KillCount:           proto.Uint32(0),
		InitShipCount:       proto.Uint32(1),
		ContinuousKillCount: proto.Uint32(0),
		MoveStepCount:       proto.Uint32(0),
	}
	stateBytes, err := proto.Marshal(current)
	if err != nil {
		t.Fatalf("failed to marshal current chapter: %v", err)
	}
	state := orm.ChapterState{CommanderID: commander.CommanderID, ChapterID: current.GetId(), State: stateBytes}
	if err := orm.UpsertChapterState(&state); err != nil {
		t.Fatalf("failed to upsert chapter state: %v", err)
	}

	buf := []byte{}
	if _, _, err := answer.ChapterBaseSync(&buf, client); err != nil {
		t.Fatalf("ChapterBaseSync failed: %v", err)
	}

	response := &protobuf.SC_13000{}
	decodeTestPacket(t, client, 13000, response)
	if response.GetCurrentChapter() == nil {
		t.Fatalf("expected current chapter")
	}
	if response.GetCurrentChapter().GetId() != current.GetId() {
		t.Fatalf("expected current chapter id %d, got %d", current.GetId(), response.GetCurrentChapter().GetId())
	}
}

func TestChapterBaseSyncEmptyStateBlob(t *testing.T) {
	if err := orm.CreateCommanderRoot(4242003, 4242003, "Chapter Base Sync 3", 0, 0); err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}
	commander, err := orm.GetCommanderCoreByID(4242003)
	if err != nil {
		t.Fatalf("failed to load commander: %v", err)
	}
	client := &connection.Client{Commander: commander}

	state := orm.ChapterState{CommanderID: commander.CommanderID, ChapterID: 0, State: []byte{}}
	if err := orm.UpsertChapterState(&state); err != nil {
		t.Fatalf("failed to upsert chapter state: %v", err)
	}

	buf := []byte{}
	if _, _, err := answer.ChapterBaseSync(&buf, client); err != nil {
		t.Fatalf("ChapterBaseSync failed: %v", err)
	}

	response := &protobuf.SC_13000{}
	decodeTestPacket(t, client, 13000, response)
	if response.GetCurrentChapter() != nil {
		t.Fatalf("expected no current chapter")
	}
}
