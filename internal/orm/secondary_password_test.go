package orm

import (
	"reflect"
	"testing"
	"time"
)

func TestSecondaryPasswordSettingsDefaults(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &SecondaryPasswordSettings{})

	settings, err := GetSecondaryPasswordSettings(GormDB, 1001)
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if settings.CommanderID != 1001 {
		t.Fatalf("expected commander id 1001, got %d", settings.CommanderID)
	}
	if settings.PasswordHash != "" || settings.Notice != "" {
		t.Fatalf("expected empty password and notice")
	}
	if len(settings.SystemList) != 0 {
		t.Fatalf("expected empty system list")
	}
	if settings.FailCount != 0 || settings.FailCd != nil {
		t.Fatalf("expected empty lockout state")
	}
}

func TestSecondaryPasswordSettingsUpsertAndLockout(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &SecondaryPasswordSettings{})

	settings := SecondaryPasswordSettings{
		CommanderID:  1002,
		PasswordHash: "hash",
		Notice:       "hint",
		SystemList:   Int64List{1, 2, 3},
		FailCount:    0,
		FailCd:       nil,
	}
	if err := UpsertSecondaryPasswordSettings(GormDB, settings); err != nil {
		t.Fatalf("upsert settings: %v", err)
	}
	stored, err := GetSecondaryPasswordSettings(GormDB, 1002)
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if stored.PasswordHash != "hash" || stored.Notice != "hint" {
		t.Fatalf("unexpected stored settings")
	}
	if !reflect.DeepEqual(stored.SystemList, Int64List{1, 2, 3}) {
		t.Fatalf("unexpected system list")
	}
	lockout := time.Now().Unix() + 60
	if err := UpdateSecondaryPasswordLockout(GormDB, 1002, 4, &lockout); err != nil {
		t.Fatalf("update lockout: %v", err)
	}
	stored, err = GetSecondaryPasswordSettings(GormDB, 1002)
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if stored.FailCount != 4 || stored.FailCd == nil {
		t.Fatalf("expected lockout values stored")
	}
	if err := ResetSecondaryPasswordLockout(GormDB, 1002); err != nil {
		t.Fatalf("reset lockout: %v", err)
	}
	stored, err = GetSecondaryPasswordSettings(GormDB, 1002)
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if stored.FailCount != 0 || stored.FailCd != nil {
		t.Fatalf("expected lockout cleared")
	}
}
