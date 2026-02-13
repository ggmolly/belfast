package orm

import "testing"

func TestRefluxStateCRUD(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &RefluxState{})

	state, err := GetOrCreateRefluxState(7001)
	if err != nil {
		t.Fatalf("get or create reflux state: %v", err)
	}
	if state.CommanderID != 7001 {
		t.Fatalf("expected commander id 7001")
	}

	state.Active = 1
	state.ReturnLv = 30
	state.ReturnTime = 1000
	state.ShipNumber = 5
	state.LastOfflineTime = 900
	state.Pt = 120
	state.SignCnt = 2
	state.SignLastTime = 1100
	state.PtStage = 3
	if err := SaveRefluxState(state); err != nil {
		t.Fatalf("save reflux state: %v", err)
	}

	loaded, err := GetOrCreateRefluxState(7001)
	if err != nil {
		t.Fatalf("load reflux state: %v", err)
	}
	if loaded.Active != 1 || loaded.ReturnLv != 30 || loaded.ReturnTime != 1000 {
		t.Fatalf("unexpected reflux state values")
	}
	if loaded.ShipNumber != 5 || loaded.LastOfflineTime != 900 || loaded.Pt != 120 {
		t.Fatalf("unexpected reflux state fields")
	}
	if loaded.SignCnt != 2 || loaded.SignLastTime != 1100 || loaded.PtStage != 3 {
		t.Fatalf("unexpected reflux progress")
	}
}
