package orm

import (
	"context"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/db"
)

func TestCommanderSurveyCompletion(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &CommanderSurvey{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 9001, AccountID: 9001, Name: "Survey Tester"}
	if err := CreateCommanderRoot(commander.CommanderID, commander.AccountID, commander.Name, 0, 0); err != nil {
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
	if err := SetCommanderSurveyCompleted(commander.CommanderID, 1001, now); err != nil {
		t.Fatalf("set survey completed: %v", err)
	}
	if err := SetCommanderSurveyCompleted(commander.CommanderID, 1001, now); err != nil {
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
	if err := db.DefaultStore.Pool.QueryRow(context.Background(), `
SELECT COUNT(*)
FROM commander_surveys
WHERE commander_id = $1
  AND survey_id = $2
`, int64(commander.CommanderID), int64(1001)).Scan(&count); err != nil {
		t.Fatalf("count survey completion: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 survey completion row, got %d", count)
	}
}
