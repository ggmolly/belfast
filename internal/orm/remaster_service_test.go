package orm

import (
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"
)

func TestRemasterStateAndReset(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &RemasterState{})

	state, err := GetOrCreateRemasterState(GormDB, 200)
	if err != nil {
		t.Fatalf("get or create remaster state: %v", err)
	}
	if state.CommanderID != 200 {
		t.Fatalf("unexpected commander id")
	}
	state.DailyCount = 5
	state.LastDailyResetAt = time.Now().Add(-24 * time.Hour)
	if !ApplyRemasterDailyReset(state, time.Now()) {
		t.Fatalf("expected daily reset")
	}
	if state.DailyCount != 0 {
		t.Fatalf("expected daily count reset")
	}
	if ApplyRemasterDailyReset(state, time.Now()) {
		t.Fatalf("expected no reset for same day")
	}

	state2, err := GetOrCreateRemasterState(GormDB, 200)
	if err != nil {
		t.Fatalf("get existing remaster state: %v", err)
	}
	if state2.CommanderID != 200 {
		t.Fatalf("unexpected commander id")
	}
}

func TestRemasterProgressCRUD(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &RemasterProgress{})

	entry := RemasterProgress{CommanderID: 201, ChapterID: 1, Pos: 1, Count: 1}
	if err := UpsertRemasterProgress(GormDB, &entry); err != nil {
		t.Fatalf("upsert progress: %v", err)
	}
	entry.Count = 2
	if err := UpsertRemasterProgress(GormDB, &entry); err != nil {
		t.Fatalf("upsert progress update: %v", err)
	}
	list, err := ListRemasterProgress(GormDB, 201)
	if err != nil || len(list) != 1 {
		t.Fatalf("list progress: %v", err)
	}
	loaded, err := GetRemasterProgress(GormDB, 201, 1, 1)
	if err != nil {
		t.Fatalf("get progress: %v", err)
	}
	if loaded.Count != 2 {
		t.Fatalf("expected count updated")
	}
	if err := DeleteRemasterProgress(GormDB, 201, 1, 1); err != nil {
		t.Fatalf("delete progress: %v", err)
	}
	_, err = GetRemasterProgress(GormDB, 201, 1, 1)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected not found after delete")
	}
}

func TestStartOfDay(t *testing.T) {
	when := time.Date(2024, time.June, 3, 12, 30, 0, 0, time.UTC)
	start := startOfDay(when)
	if start.Hour() != 0 || start.Minute() != 0 || start.Day() != 3 {
		t.Fatalf("unexpected start of day: %v", start)
	}
}
