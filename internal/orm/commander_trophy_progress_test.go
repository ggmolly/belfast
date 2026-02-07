package orm

import (
	"testing"
)

func TestGetOrCreateCommanderTrophyProgressCreatesAndReuses(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &CommanderTrophyProgress{})

	row, created, err := GetOrCreateCommanderTrophyProgress(GormDB, 1, 100, 5)
	if err != nil {
		t.Fatalf("create trophy progress: %v", err)
	}
	if !created {
		t.Fatalf("expected created")
	}
	if row.Progress != 5 || row.Timestamp != 0 {
		t.Fatalf("unexpected row values")
	}

	row2, created2, err := GetOrCreateCommanderTrophyProgress(GormDB, 1, 100, 99)
	if err != nil {
		t.Fatalf("get trophy progress: %v", err)
	}
	if created2 {
		t.Fatalf("expected not created")
	}
	if row2.Progress != 5 {
		t.Fatalf("expected progress unchanged, got %d", row2.Progress)
	}
}

func TestClaimCommanderTrophyProgressUpdatesTimestamp(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &CommanderTrophyProgress{})

	if err := GormDB.Create(&CommanderTrophyProgress{CommanderID: 2, TrophyID: 200, Progress: 1, Timestamp: 0}).Error; err != nil {
		t.Fatalf("seed trophy progress: %v", err)
	}
	if err := ClaimCommanderTrophyProgress(GormDB, 2, 200, 1234); err != nil {
		t.Fatalf("claim trophy: %v", err)
	}
	stored, err := GetCommanderTrophyProgress(GormDB, 2, 200)
	if err != nil {
		t.Fatalf("load trophy: %v", err)
	}
	if stored.Timestamp != 1234 {
		t.Fatalf("expected timestamp 1234, got %d", stored.Timestamp)
	}
}
