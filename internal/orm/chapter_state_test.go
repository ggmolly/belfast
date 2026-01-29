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

func TestUpsertAndDeleteChapterState(t *testing.T) {
	initBattleSessionTestDB(t)
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&ChapterState{}).Error; err != nil {
		t.Fatalf("clear chapter state: %v", err)
	}
	state := ChapterState{CommanderID: 9100, ChapterID: 202, State: []byte{7}}
	if err := UpsertChapterState(GormDB, &state); err != nil {
		t.Fatalf("upsert chapter state: %v", err)
	}
	fetched, err := GetChapterState(GormDB, 9100)
	if err != nil {
		t.Fatalf("get chapter state: %v", err)
	}
	if fetched.ChapterID != 202 || fetched.State[0] != 7 {
		t.Fatalf("unexpected state: %+v", fetched)
	}
	state.State = []byte{8}
	if err := UpsertChapterState(GormDB, &state); err != nil {
		t.Fatalf("upsert chapter state update: %v", err)
	}
	fetched, err = GetChapterState(GormDB, 9100)
	if err != nil {
		t.Fatalf("get chapter state after update: %v", err)
	}
	if fetched.State[0] != 8 {
		t.Fatalf("expected updated state")
	}
	if err := DeleteChapterState(GormDB, 9100); err != nil {
		t.Fatalf("delete chapter state: %v", err)
	}
	_, err = GetChapterState(GormDB, 9100)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected record not found after delete, got %v", err)
	}
}
