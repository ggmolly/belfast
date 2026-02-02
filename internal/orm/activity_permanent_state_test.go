package orm

import (
	"os"
	"testing"

	"gorm.io/gorm"
)

func clearPermanentStateTable(t *testing.T) {
	t.Helper()
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&ActivityPermanentState{}).Error; err != nil {
		t.Fatalf("failed to clear activity permanent state: %v", err)
	}
}

func TestActivityPermanentStateCreateAndUpdate(t *testing.T) {
	os.Setenv("MODE", "test")
	InitDatabase()
	clearPermanentStateTable(t)

	state, err := GetOrCreateActivityPermanentState(GormDB, 1)
	if err != nil {
		t.Fatalf("create permanent state failed: %v", err)
	}
	if state.CurrentActivityID != 0 {
		t.Fatalf("expected current activity to default to 0")
	}
	if len(state.FinishedList()) != 0 {
		t.Fatalf("expected finished list to be empty")
	}

	state.CurrentActivityID = 6000
	state.AddFinished(6000)
	state.AddFinished(6000)
	if err := SaveActivityPermanentState(GormDB, state); err != nil {
		t.Fatalf("save permanent state failed: %v", err)
	}

	reloaded, err := GetOrCreateActivityPermanentState(GormDB, 1)
	if err != nil {
		t.Fatalf("reload permanent state failed: %v", err)
	}
	if reloaded.CurrentActivityID != 6000 {
		t.Fatalf("expected current activity to be 6000")
	}
	if len(reloaded.FinishedList()) != 1 || reloaded.FinishedList()[0] != 6000 {
		t.Fatalf("expected finished list to include 6000")
	}
}
