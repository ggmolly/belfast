package orm

import "testing"

func TestSecondaryPasswordStateCRUD(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &SecondaryPasswordState{})

	state, err := GetOrCreateSecondaryPasswordState(GormDB, 9001)
	if err != nil {
		t.Fatalf("get or create secondary password state: %v", err)
	}
	if state.CommanderID != 9001 {
		t.Fatalf("expected commander id 9001")
	}
	if len(state.SystemList) != 0 {
		t.Fatalf("expected empty system list")
	}

	state.PasswordHash = "hash"
	state.Notice = "notice"
	state.SystemList = Int64List{1, 2}
	state.State = 1
	state.FailCount = 2
	state.FailCd = 123
	if err := SaveSecondaryPasswordState(GormDB, state); err != nil {
		t.Fatalf("save secondary password state: %v", err)
	}

	loaded, err := GetOrCreateSecondaryPasswordState(GormDB, 9001)
	if err != nil {
		t.Fatalf("load secondary password state: %v", err)
	}
	if loaded.PasswordHash != "hash" || loaded.Notice != "notice" {
		t.Fatalf("expected stored hash and notice")
	}
	if len(loaded.SystemList) != 2 || loaded.SystemList[0] != 1 || loaded.SystemList[1] != 2 {
		t.Fatalf("unexpected system list: %v", loaded.SystemList)
	}
	if loaded.State != 1 || loaded.FailCount != 2 || loaded.FailCd != 123 {
		t.Fatalf("unexpected state values")
	}
}
