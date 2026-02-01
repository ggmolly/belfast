package orm

import (
	"testing"
	"time"
)

func TestCommanderSurveyCompletion(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &CommanderSurvey{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 9001, AccountID: 9001, Name: "Survey Tester"}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}

	completed, err := IsCommanderSurveyCompleted(commander.CommanderID, 1001)
	if err != nil {
		t.Fatalf("check survey completion: %v", err)
	}
	if completed {
		t.Fatalf("expected survey not completed")
	}

	now := time.Now().UTC()
	if err := SetCommanderSurveyCompleted(GormDB, commander.CommanderID, 1001, now); err != nil {
		t.Fatalf("set survey completed: %v", err)
	}
	if err := SetCommanderSurveyCompleted(GormDB, commander.CommanderID, 1001, now); err != nil {
		t.Fatalf("set survey completed again: %v", err)
	}

	completed, err = IsCommanderSurveyCompleted(commander.CommanderID, 1001)
	if err != nil {
		t.Fatalf("check survey completion: %v", err)
	}
	if !completed {
		t.Fatalf("expected survey completed")
	}

	var count int64
	if err := GormDB.Model(&CommanderSurvey{}).Where("commander_id = ? AND survey_id = ?", commander.CommanderID, 1001).Count(&count).Error; err != nil {
		t.Fatalf("count survey completion: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 survey completion row, got %d", count)
	}
}
