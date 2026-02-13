package orm

import (
	"testing"
	"time"
)

func TestListCommanderActiveBuffs(t *testing.T) {
	initRandomFlagShipTestDB(t)
	commanderID := uint32(1010)
	otherCommanderID := uint32(2020)
	now := time.Date(2026, 1, 22, 12, 0, 0, 0, time.UTC)

	clearTable(t, &CommanderBuff{})

	entries := []CommanderBuff{
		{CommanderID: commanderID, BuffID: 10, ExpiresAt: now.Add(-time.Hour)},
		{CommanderID: commanderID, BuffID: 11, ExpiresAt: now.Add(time.Hour)},
		{CommanderID: otherCommanderID, BuffID: 12, ExpiresAt: now.Add(time.Hour)},
	}
	for i := range entries {
		if err := UpsertCommanderBuff(entries[i].CommanderID, entries[i].BuffID, entries[i].ExpiresAt); err != nil {
			t.Fatalf("create commander buffs: %v", err)
		}
	}

	active, err := ListCommanderActiveBuffs(commanderID, now)
	if err != nil {
		t.Fatalf("list commander buffs: %v", err)
	}
	if len(active) != 1 {
		t.Fatalf("expected 1 active buff, got %d", len(active))
	}
	if active[0].BuffID != 11 {
		t.Fatalf("expected buff id 11, got %d", active[0].BuffID)
	}
}
