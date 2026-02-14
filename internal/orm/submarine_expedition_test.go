package orm

import (
	"errors"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
)

func TestSubmarineExpeditionStateUpsertCreatesRow(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &SubmarineExpeditionState{})

	state := SubmarineExpeditionState{
		CommanderID:        1,
		LastRefreshTime:    123,
		WeeklyRefreshCount: 2,
		ActiveChapterID:    1000,
		OverallProgress:    7,
	}
	if err := UpsertSubmarineState(&state); err != nil {
		t.Fatalf("upsert state: %v", err)
	}

	stored, err := GetSubmarineState(1)
	if err != nil {
		t.Fatalf("get state: %v", err)
	}
	if stored.CommanderID != 1 || stored.ActiveChapterID != 1000 {
		t.Fatalf("unexpected stored state")
	}
}

func TestGetSubmarineStateReturnsRecordNotFound(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &SubmarineExpeditionState{})

	_, err := GetSubmarineState(999)
	if !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("expected record not found")
	}
}

func TestResetWeeklyRefreshClearsWeeklyRefreshCount(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &SubmarineExpeditionState{})

	state := SubmarineExpeditionState{CommanderID: 2, WeeklyRefreshCount: 3, LastRefreshTime: 1}
	if err := UpsertSubmarineState(&state); err != nil {
		t.Fatalf("seed state: %v", err)
	}
	if err := ResetWeeklyRefresh(2, 555); err != nil {
		t.Fatalf("reset weekly refresh: %v", err)
	}
	stored, err := GetSubmarineState(2)
	if err != nil {
		t.Fatalf("get state: %v", err)
	}
	if stored.WeeklyRefreshCount != 0 {
		t.Fatalf("expected weekly refresh count 0, got %d", stored.WeeklyRefreshCount)
	}
	if stored.LastRefreshTime != 555 {
		t.Fatalf("expected last refresh time 555, got %d", stored.LastRefreshTime)
	}
}
