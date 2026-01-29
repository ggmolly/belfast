package orm

import (
	"errors"
	"testing"

	"gorm.io/gorm"
)

func TestUpsertChapterProgressCreatesAndUpdates(t *testing.T) {
	initBattleSessionTestDB(t)
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&ChapterProgress{}).Error; err != nil {
		t.Fatalf("clear chapter progress: %v", err)
	}
	progress := ChapterProgress{
		CommanderID:      1000,
		ChapterID:        101,
		Progress:         0,
		KillBossCount:    1,
		KillEnemyCount:   2,
		TakeBoxCount:     3,
		DefeatCount:      1,
		TodayDefeatCount: 1,
		PassCount:        0,
	}
	if err := UpsertChapterProgress(GormDB, &progress); err != nil {
		t.Fatalf("upsert progress: %v", err)
	}
	stored, err := GetChapterProgress(GormDB, 1000, 101)
	if err != nil {
		t.Fatalf("get progress: %v", err)
	}
	if stored.KillBossCount != 1 || stored.KillEnemyCount != 2 || stored.TakeBoxCount != 3 {
		t.Fatalf("unexpected stored counts")
	}

	progress.KillEnemyCount = 5
	progress.Progress = 100
	progress.PassCount = 1
	if err := UpsertChapterProgress(GormDB, &progress); err != nil {
		t.Fatalf("upsert progress update: %v", err)
	}
	stored, err = GetChapterProgress(GormDB, 1000, 101)
	if err != nil {
		t.Fatalf("get progress after update: %v", err)
	}
	if stored.KillEnemyCount != 5 || stored.Progress != 100 || stored.PassCount != 1 {
		t.Fatalf("unexpected updated progress values")
	}
}

func TestDeleteChapterProgress(t *testing.T) {
	initBattleSessionTestDB(t)
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&ChapterProgress{}).Error; err != nil {
		t.Fatalf("clear chapter progress: %v", err)
	}
	progress := ChapterProgress{
		CommanderID: 2000,
		ChapterID:   202,
	}
	if err := UpsertChapterProgress(GormDB, &progress); err != nil {
		t.Fatalf("upsert progress: %v", err)
	}
	if err := DeleteChapterProgress(GormDB, 2000, 202); err != nil {
		t.Fatalf("delete progress: %v", err)
	}
	_, err := GetChapterProgress(GormDB, 2000, 202)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected record not found after delete, got %v", err)
	}
}

func TestListChapterProgressOrdersByChapter(t *testing.T) {
	initBattleSessionTestDB(t)
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&ChapterProgress{}).Error; err != nil {
		t.Fatalf("clear chapter progress: %v", err)
	}
	entries := []ChapterProgress{
		{CommanderID: 3000, ChapterID: 3, Progress: 1},
		{CommanderID: 3000, ChapterID: 1, Progress: 2},
		{CommanderID: 3000, ChapterID: 2, Progress: 3},
	}
	for i := range entries {
		if err := UpsertChapterProgress(GormDB, &entries[i]); err != nil {
			t.Fatalf("upsert progress: %v", err)
		}
	}
	list, err := ListChapterProgress(GormDB, 3000)
	if err != nil {
		t.Fatalf("list progress: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(list))
	}
	if list[0].ChapterID != 1 || list[1].ChapterID != 2 || list[2].ChapterID != 3 {
		t.Fatalf("unexpected order: %+v", list)
	}
}
