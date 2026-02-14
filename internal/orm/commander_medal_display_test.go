package orm

import (
	"testing"
)

func TestCommanderMedalDisplayPersistsOrderedList(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &CommanderMedalDisplay{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 9001, AccountID: 9001, Name: "Medal Display"}
	if err := CreateCommanderRoot(commander.CommanderID, commander.AccountID, commander.Name, 0, 0); err != nil {
		t.Fatalf("seed commander: %v", err)
	}

	medals := []uint32{10, 20, 30, 40, 50}
	if err := SetCommanderMedalDisplay(commander.CommanderID, medals); err != nil {
		t.Fatalf("set medal display: %v", err)
	}
	stored, err := ListCommanderMedalDisplay(commander.CommanderID)
	if err != nil {
		t.Fatalf("list medal display: %v", err)
	}
	if len(stored) != len(medals) {
		t.Fatalf("expected %d medals, got %d", len(medals), len(stored))
	}
	for i := range medals {
		if stored[i] != medals[i] {
			t.Fatalf("expected medal %d at %d, got %d", medals[i], i, stored[i])
		}
	}
}

func TestCommanderMedalDisplayReplaceIsAtomic(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &CommanderMedalDisplay{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 9002, AccountID: 9002, Name: "Medal Replace"}
	if err := CreateCommanderRoot(commander.CommanderID, commander.AccountID, commander.Name, 0, 0); err != nil {
		t.Fatalf("seed commander: %v", err)
	}

	if err := SetCommanderMedalDisplay(commander.CommanderID, []uint32{1, 2, 3}); err != nil {
		t.Fatalf("seed medal display: %v", err)
	}
	if err := SetCommanderMedalDisplay(commander.CommanderID, []uint32{100, 200}); err != nil {
		t.Fatalf("replace medal display: %v", err)
	}
	stored, err := ListCommanderMedalDisplay(commander.CommanderID)
	if err != nil {
		t.Fatalf("list medal display: %v", err)
	}
	if len(stored) != 2 {
		t.Fatalf("expected 2 medals, got %d", len(stored))
	}
	if stored[0] != 100 || stored[1] != 200 {
		t.Fatalf("expected [100 200], got %v", stored)
	}
}
