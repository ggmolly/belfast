package answer

import (
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
)

func TestCommanderStoryProgressDisplaysThreeStarsAfterClears(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.ChapterProgress{})
	seedConfigEntry(t, "sharecfgdata/chapter_template.json", "101", `{"id":101,"star_require_1":1,"num_1":1,"star_require_2":2,"num_2":3,"star_require_3":3,"num_3":1}`)

	progress := &orm.ChapterProgress{
		CommanderID:    client.Commander.CommanderID,
		ChapterID:      101,
		Progress:       100,
		KillBossCount:  0,
		KillEnemyCount: 1,
		TakeBoxCount:   0,
		PassCount:      3,
	}
	if err := orm.UpsertChapterProgress(orm.GormDB, progress); err != nil {
		t.Fatalf("seed chapter progress: %v", err)
	}

	buffer := []byte{}
	if _, _, err := CommanderStoryProgress(&buffer, client); err != nil {
		t.Fatalf("commander story progress failed: %v", err)
	}
	var response protobuf.SC_13001
	decodeResponse(t, client, &response)
	if len(response.GetChapterList()) != 1 {
		t.Fatalf("expected 1 chapter entry, got %d", len(response.GetChapterList()))
	}
	entry := response.GetChapterList()[0]
	if entry.GetKillBossCount() != 1 {
		t.Fatalf("expected kill boss count 1, got %d", entry.GetKillBossCount())
	}
	if entry.GetKillEnemyCount() != 3 {
		t.Fatalf("expected kill enemy count 3, got %d", entry.GetKillEnemyCount())
	}
	if entry.GetTakeBoxCount() != 1 {
		t.Fatalf("expected take box count 1, got %d", entry.GetTakeBoxCount())
	}
	stored, err := orm.GetChapterProgress(orm.GormDB, client.Commander.CommanderID, 101)
	if err != nil {
		t.Fatalf("load chapter progress: %v", err)
	}
	if stored.KillBossCount != 0 {
		t.Fatalf("expected stored kill boss count 0, got %d", stored.KillBossCount)
	}
	if stored.KillEnemyCount != 1 {
		t.Fatalf("expected stored kill enemy count 1, got %d", stored.KillEnemyCount)
	}
	if stored.TakeBoxCount != 0 {
		t.Fatalf("expected stored take box count 0, got %d", stored.TakeBoxCount)
	}
}

func TestCommanderStoryProgressUsesStoredRemasterActiveChapter(t *testing.T) {
	client := setupPlayerUpdateTest(t)
	clearTable(t, &orm.RemasterState{})
	state := orm.RemasterState{
		CommanderID:      client.Commander.CommanderID,
		ActiveChapterID:  77,
		LastDailyResetAt: time.Now(),
	}
	if err := orm.GormDB.Create(&state).Error; err != nil {
		t.Fatalf("seed remaster state: %v", err)
	}

	buffer := []byte{}
	if _, _, err := CommanderStoryProgress(&buffer, client); err != nil {
		t.Fatalf("commander story progress failed: %v", err)
	}
	var response protobuf.SC_13001
	decodeResponse(t, client, &response)
	if response.GetReactChapter().GetActiveId() != 77 {
		t.Fatalf("expected active id 77, got %d", response.GetReactChapter().GetActiveId())
	}
}
