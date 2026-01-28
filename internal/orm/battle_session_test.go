package orm

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	"gorm.io/gorm"
)

var battleSessionTestOnce sync.Once

func initBattleSessionTestDB(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	battleSessionTestOnce.Do(func() {
		InitDatabase()
	})
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&BattleSession{}).Error; err != nil {
		t.Fatalf("clear battle sessions: %v", err)
	}
}

func TestUpsertBattleSessionCreatesAndUpdates(t *testing.T) {
	initBattleSessionTestDB(t)
	session := BattleSession{
		CommanderID: 100,
		System:      1,
		StageID:     200,
		Key:         333,
		ShipIDs:     Int64List{101, 102},
	}
	if err := UpsertBattleSession(GormDB, &session); err != nil {
		t.Fatalf("upsert session: %v", err)
	}
	stored, err := GetBattleSession(GormDB, 100)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if stored.System != 1 || stored.StageID != 200 || stored.Key != 333 {
		t.Fatalf("unexpected stored session values")
	}
	if !reflect.DeepEqual(stored.ShipIDs, Int64List{101, 102}) {
		t.Fatalf("unexpected ship ids: %v", stored.ShipIDs)
	}

	session.System = 2
	session.StageID = 300
	session.Key = 444
	session.ShipIDs = Int64List{201}
	if err := UpsertBattleSession(GormDB, &session); err != nil {
		t.Fatalf("upsert session update: %v", err)
	}
	stored, err = GetBattleSession(GormDB, 100)
	if err != nil {
		t.Fatalf("get session after update: %v", err)
	}
	if stored.System != 2 || stored.StageID != 300 || stored.Key != 444 {
		t.Fatalf("unexpected updated session values")
	}
	if !reflect.DeepEqual(stored.ShipIDs, Int64List{201}) {
		t.Fatalf("unexpected updated ship ids: %v", stored.ShipIDs)
	}
}

func TestDeleteBattleSession(t *testing.T) {
	initBattleSessionTestDB(t)
	session := BattleSession{
		CommanderID: 200,
		System:      1,
		StageID:     500,
		Key:         777,
		ShipIDs:     Int64List{},
	}
	if err := UpsertBattleSession(GormDB, &session); err != nil {
		t.Fatalf("upsert session: %v", err)
	}
	if err := DeleteBattleSession(GormDB, 200); err != nil {
		t.Fatalf("delete session: %v", err)
	}
	_, err := GetBattleSession(GormDB, 200)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected record not found after delete, got %v", err)
	}
}
