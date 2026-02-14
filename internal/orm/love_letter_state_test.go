package orm

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
)

var loveLetterStateTestCommanderSeed uint32 = 9800

func nextLoveLetterStateCommanderID() uint32 {
	return atomic.AddUint32(&loveLetterStateTestCommanderSeed, 1)
}

func TestCommanderLoveLetterStateCRUD(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &CommanderLoveLetterState{})
	commanderID := nextLoveLetterStateCommanderID()
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), fmt.Sprintf("DELETE FROM %s WHERE commander_id = $1", QualifiedTable("commanders")), int64(commanderID)); err != nil {
		t.Fatalf("delete commander: %v", err)
	}
	if err := CreateCommanderRoot(commanderID, commanderID, "LoveLetter ORM Tester", 0, 0); err != nil {
		t.Fatalf("create commander root: %v", err)
	}

	state, err := GetOrCreateCommanderLoveLetterState(commanderID)
	if err != nil {
		t.Fatalf("get or create love letter state: %v", err)
	}
	if state.CommanderID != commanderID {
		t.Fatalf("unexpected commander id %d", state.CommanderID)
	}
	if len(state.Medals) != 0 || len(state.ManualLetters) != 0 || len(state.ConvertedItems) != 0 || len(state.RewardedIDs) != 0 {
		t.Fatalf("expected empty initial state, got %+v", state)
	}
	if len(state.LetterContents) != 0 {
		t.Fatalf("expected empty initial letter contents")
	}

	state.Medals = []LoveLetterMedalState{{GroupID: 10000, Exp: 30, Level: 2}}
	state.ManualLetters = []LoveLetterLetterState{{GroupID: 10000, LetterIDList: []uint32{2018001, 2019001}}}
	state.ConvertedItems = []LoveLetterConvertedItem{{ItemID: 41002, GroupID: 10000, Year: 2018}}
	state.RewardedIDs = []uint32{1, 3}
	state.LetterContents = map[uint32]string{2018001: "dear commander"}
	if err := SaveCommanderLoveLetterState(state); err != nil {
		t.Fatalf("save state: %v", err)
	}

	loaded, err := GetCommanderLoveLetterState(commanderID)
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	if len(loaded.Medals) != 1 || loaded.Medals[0].Level != 2 {
		t.Fatalf("unexpected medals: %+v", loaded.Medals)
	}
	if len(loaded.ManualLetters) != 1 || len(loaded.ManualLetters[0].LetterIDList) != 2 {
		t.Fatalf("unexpected manual letters: %+v", loaded.ManualLetters)
	}
	if len(loaded.ConvertedItems) != 1 || loaded.ConvertedItems[0].ItemID != 41002 {
		t.Fatalf("unexpected converted items: %+v", loaded.ConvertedItems)
	}
	if len(loaded.RewardedIDs) != 2 || loaded.RewardedIDs[1] != 3 {
		t.Fatalf("unexpected rewarded ids: %+v", loaded.RewardedIDs)
	}
	if loaded.LetterContents[2018001] != "dear commander" {
		t.Fatalf("unexpected letter contents: %+v", loaded.LetterContents)
	}

	if err := DeleteCommanderLoveLetterState(commanderID); err != nil {
		t.Fatalf("delete state: %v", err)
	}
	_, err = GetCommanderLoveLetterState(commanderID)
	if !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("expected not found after delete, got %v", err)
	}
}
