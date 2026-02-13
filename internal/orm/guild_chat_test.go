package orm

import (
	"context"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/db"
)

func TestGuildChatMessageHistoryOrder(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &GuildChatMessage{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 100, AccountID: 100, Name: "Guild Chatter"}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commanders (commander_id, account_id, name) VALUES ($1, $2, $3)`, int64(commander.CommanderID), int64(commander.AccountID), commander.Name); err != nil {
		t.Fatalf("create commander: %v", err)
	}

	base := time.Date(2026, time.January, 1, 10, 0, 0, 0, time.UTC)
	if _, err := CreateGuildChatMessage(0, commander.CommanderID, "first", base); err != nil {
		t.Fatalf("create message 1: %v", err)
	}
	if _, err := CreateGuildChatMessage(0, commander.CommanderID, "second", base.Add(2*time.Minute)); err != nil {
		t.Fatalf("create message 2: %v", err)
	}
	if _, err := CreateGuildChatMessage(0, commander.CommanderID, "third", base.Add(4*time.Minute)); err != nil {
		t.Fatalf("create message 3: %v", err)
	}

	messages, err := ListGuildChatMessages(0, 2)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}
	if messages[0].Content != "second" || messages[1].Content != "third" {
		t.Fatalf("unexpected order: %v, %v", messages[0].Content, messages[1].Content)
	}
	if messages[0].Sender.CommanderID != commander.CommanderID {
		t.Fatalf("expected sender preload")
	}
}
