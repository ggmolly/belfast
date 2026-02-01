package orm

import "testing"

func TestUpsertSurveyStateCreatesAndUpdates(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &SurveyState{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 1201, AccountID: 1, Name: "Survey Tester"}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}

	first := SurveyState{CommanderID: commander.CommanderID, SurveyID: 1001}
	if err := UpsertSurveyState(GormDB, &first); err != nil {
		t.Fatalf("upsert first: %v", err)
	}
	second := SurveyState{CommanderID: commander.CommanderID, SurveyID: 1002}
	if err := UpsertSurveyState(GormDB, &second); err != nil {
		t.Fatalf("upsert second: %v", err)
	}
	stored, err := GetSurveyState(GormDB, commander.CommanderID)
	if err != nil {
		t.Fatalf("get state: %v", err)
	}
	if stored.SurveyID != 1002 {
		t.Fatalf("expected survey id 1002, got %d", stored.SurveyID)
	}
}
