package orm

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/db"
)

func TestGetChapterStateExpiresAfter24h(t *testing.T) {
	initBattleSessionTestDB(t)
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `DELETE FROM chapter_states`); err != nil {
		t.Fatalf("clear chapter state: %v", err)
	}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `DELETE FROM chapter_progress`); err != nil {
		t.Fatalf("clear chapter progress: %v", err)
	}
	state := ChapterState{
		CommanderID: 9001,
		ChapterID:   101,
		State:       []byte{1},
		UpdatedAt:   uint32(time.Now().Unix()) - 60*60*25,
	}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `
INSERT INTO chapter_states (commander_id, chapter_id, state, updated_at)
VALUES ($1, $2, $3, $4)
`, int64(state.CommanderID), int64(state.ChapterID), state.State, int64(state.UpdatedAt)); err != nil {
		t.Fatalf("seed chapter state: %v", err)
	}
	_, err := GetChapterState(state.CommanderID)
	if !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("expected record not found, got %v", err)
	}
	var count int64
	if err := db.DefaultStore.Pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM chapter_states WHERE commander_id = $1`, int64(state.CommanderID)).Scan(&count); err != nil {
		t.Fatalf("count chapter state: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected expired chapter state to be deleted")
	}
}

func TestGetChapterStateRetainsCompletedChapters(t *testing.T) {
	initBattleSessionTestDB(t)
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `DELETE FROM chapter_states`); err != nil {
		t.Fatalf("clear chapter state: %v", err)
	}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `DELETE FROM chapter_progress`); err != nil {
		t.Fatalf("clear chapter progress: %v", err)
	}
	state := ChapterState{
		CommanderID: 9002,
		ChapterID:   101,
		State:       []byte{1},
		UpdatedAt:   uint32(time.Now().Unix()) - 60*60*25,
	}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `
INSERT INTO chapter_states (commander_id, chapter_id, state, updated_at)
VALUES ($1, $2, $3, $4)
`, int64(state.CommanderID), int64(state.ChapterID), state.State, int64(state.UpdatedAt)); err != nil {
		t.Fatalf("seed chapter state: %v", err)
	}
	progress := ChapterProgress{
		CommanderID: 9002,
		ChapterID:   101,
		Progress:    100,
	}
	if err := UpsertChapterProgress(&progress); err != nil {
		t.Fatalf("seed chapter progress: %v", err)
	}
	if _, err := GetChapterState(state.CommanderID); err != nil {
		t.Fatalf("expected chapter state retained, got %v", err)
	}
}

func TestUpsertAndDeleteChapterState(t *testing.T) {
	initBattleSessionTestDB(t)
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `DELETE FROM chapter_states`); err != nil {
		t.Fatalf("clear chapter state: %v", err)
	}
	state := ChapterState{CommanderID: 9100, ChapterID: 202, State: []byte{7}}
	if err := UpsertChapterState(&state); err != nil {
		t.Fatalf("upsert chapter state: %v", err)
	}
	fetched, err := GetChapterState(9100)
	if err != nil {
		t.Fatalf("get chapter state: %v", err)
	}
	if fetched.ChapterID != 202 || fetched.State[0] != 7 {
		t.Fatalf("unexpected state: %+v", fetched)
	}
	state.State = []byte{8}
	if err := UpsertChapterState(&state); err != nil {
		t.Fatalf("upsert chapter state update: %v", err)
	}
	fetched, err = GetChapterState(9100)
	if err != nil {
		t.Fatalf("get chapter state after update: %v", err)
	}
	if fetched.State[0] != 8 {
		t.Fatalf("expected updated state")
	}
	if err := DeleteChapterState(9100); err != nil {
		t.Fatalf("delete chapter state: %v", err)
	}
	_, err = GetChapterState(9100)
	if !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("expected record not found after delete, got %v", err)
	}
}

func TestSearchChapterStatesUsesDBPaginationAndTotal(t *testing.T) {
	initBattleSessionTestDB(t)
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `DELETE FROM chapter_states`); err != nil {
		t.Fatalf("clear chapter states: %v", err)
	}

	state := ChapterState{CommanderID: 9300, ChapterID: 303, State: []byte{1}}
	if err := UpsertChapterState(&state); err != nil {
		t.Fatalf("upsert chapter state: %v", err)
	}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `
UPDATE chapter_states
SET updated_at = $2
WHERE commander_id = $1
`, int64(9300), int64(500)); err != nil {
		t.Fatalf("set updated_at: %v", err)
	}

	result, err := SearchChapterStates(9300, nil, nil, 1, 1)
	if err != nil {
		t.Fatalf("search chapter states: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected total 1, got %d", result.Total)
	}
	if len(result.States) != 0 {
		t.Fatalf("expected empty page, got %d rows", len(result.States))
	}

	updatedSince := uint32(400)
	filtered, err := SearchChapterStates(9300, nil, &updatedSince, 0, 0)
	if err != nil {
		t.Fatalf("search chapter states updated_since: %v", err)
	}
	if filtered.Total != 1 || len(filtered.States) != 1 {
		t.Fatalf("unexpected updated_since result: %+v", filtered)
	}

	chapterID := uint32(999)
	empty, err := SearchChapterStates(9300, &chapterID, nil, 0, 0)
	if err != nil {
		t.Fatalf("search chapter states chapter_id: %v", err)
	}
	if empty.Total != 0 || len(empty.States) != 0 {
		t.Fatalf("expected no rows for non-matching chapter_id, got %+v", empty)
	}
}
