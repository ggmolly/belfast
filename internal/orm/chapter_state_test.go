package orm

import (
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"
)

func TestGetChapterStateExpiresAfter24h(t *testing.T) {
	initBattleSessionTestDB(t)
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&ChapterState{}).Error; err != nil {
		t.Fatalf("clear chapter state: %v", err)
	}
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&ChapterProgress{}).Error; err != nil {
		t.Fatalf("clear chapter progress: %v", err)
	}
	state := ChapterState{
		CommanderID: 9001,
		ChapterID:   101,
		State:       []byte{1},
		UpdatedAt:   uint32(time.Now().Unix()) - 60*60*25,
	}
	if err := GormDB.Create(&state).Error; err != nil {
		t.Fatalf("seed chapter state: %v", err)
	}
	_, err := GetChapterState(GormDB, state.CommanderID)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected record not found, got %v", err)
	}
	var count int64
	if err := GormDB.Model(&ChapterState{}).Where("commander_id = ?", state.CommanderID).Count(&count).Error; err != nil {
		t.Fatalf("count chapter state: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected expired chapter state to be deleted")
	}
}

func TestGetChapterStateRetainsCompletedChapters(t *testing.T) {
	initBattleSessionTestDB(t)
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&ChapterState{}).Error; err != nil {
		t.Fatalf("clear chapter state: %v", err)
	}
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&ChapterProgress{}).Error; err != nil {
		t.Fatalf("clear chapter progress: %v", err)
	}
	state := ChapterState{
		CommanderID: 9002,
		ChapterID:   101,
		State:       []byte{1},
		UpdatedAt:   uint32(time.Now().Unix()) - 60*60*25,
	}
	if err := GormDB.Create(&state).Error; err != nil {
		t.Fatalf("seed chapter state: %v", err)
	}
	progress := ChapterProgress{
		CommanderID: 9002,
		ChapterID:   101,
		Progress:    100,
	}
	if err := UpsertChapterProgress(GormDB, &progress); err != nil {
		t.Fatalf("seed chapter progress: %v", err)
	}
	if _, err := GetChapterState(GormDB, state.CommanderID); err != nil {
		t.Fatalf("expected chapter state retained, got %v", err)
	}
}
