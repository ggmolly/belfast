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
