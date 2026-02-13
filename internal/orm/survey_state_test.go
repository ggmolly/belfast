package orm

import "testing"

func TestUpsertSurveyStateCreatesAndUpdates(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &SurveyState{})
	clearTable(t, &Commander{})

	commanderID := uint32(1201)
	if err := CreateCommanderRoot(commanderID, 1, "Survey Tester", 0, 0); err != nil {
		t.Fatalf("create commander: %v", err)
	}

	first := SurveyState{CommanderID: commanderID, SurveyID: 1001}
	if err := UpsertSurveyState(&first); err != nil {
		t.Fatalf("upsert first: %v", err)
	}
	second := SurveyState{CommanderID: commanderID, SurveyID: 1002}
	if err := UpsertSurveyState(&second); err != nil {
		t.Fatalf("upsert second: %v", err)
	}
	stored, err := GetSurveyState(commanderID)
	if err != nil {
		t.Fatalf("get state: %v", err)
	}
	if stored.SurveyID != 1002 {
		t.Fatalf("expected survey id 1002, got %d", stored.SurveyID)
	}
}
