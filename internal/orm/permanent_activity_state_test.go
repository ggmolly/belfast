package orm

import (
	"reflect"
	"sync"
	"testing"

	"gorm.io/gorm"
)

var permanentActivityStateOnce sync.Once

func initPermanentActivityStateTestDB(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	permanentActivityStateOnce.Do(func() {
		InitDatabase()
	})
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&PermanentActivityState{}).Error; err != nil {
		t.Fatalf("clear permanent activity states: %v", err)
	}
}

func TestGetOrCreatePermanentActivityState(t *testing.T) {
	initPermanentActivityStateTestDB(t)
	state, err := GetOrCreatePermanentActivityState(GormDB, 10)
	if err != nil {
		t.Fatalf("get or create failed: %v", err)
	}
	if state.PermanentNow != 0 {
		t.Fatalf("expected permanent now 0, got %d", state.PermanentNow)
	}
	if len(state.FinishedActivityIDs) != 0 {
		t.Fatalf("expected empty finished list")
	}
}

func TestPermanentActivityStateSaveFinished(t *testing.T) {
	initPermanentActivityStateTestDB(t)
	state, err := GetOrCreatePermanentActivityState(GormDB, 20)
	if err != nil {
		t.Fatalf("get or create failed: %v", err)
	}
	state.FinishedActivityIDs = ToInt64List([]uint32{6000, 6001})
	if err := GormDB.Save(state).Error; err != nil {
		t.Fatalf("save failed: %v", err)
	}
	loaded, err := GetOrCreatePermanentActivityState(GormDB, 20)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	if !reflect.DeepEqual(ToUint32List(loaded.FinishedActivityIDs), []uint32{6000, 6001}) {
		t.Fatalf("expected finished list to be persisted")
	}
}
