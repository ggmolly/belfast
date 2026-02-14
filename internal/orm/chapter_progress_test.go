package orm

import (
	"context"
	"errors"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
)

func TestUpsertChapterProgressCreatesAndUpdates(t *testing.T) {
	initBattleSessionTestDB(t)
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `DELETE FROM chapter_progress`); err != nil {
		t.Fatalf("clear chapter progress: %v", err)
	}
	progress := ChapterProgress{
		CommanderID:      1000,
		ChapterID:        101,
		Progress:         0,
		KillBossCount:    1,
		KillEnemyCount:   2,
		TakeBoxCount:     3,
		DefeatCount:      1,
		TodayDefeatCount: 1,
		PassCount:        0,
	}
	if err := UpsertChapterProgress(&progress); err != nil {
		t.Fatalf("upsert progress: %v", err)
	}
	stored, err := GetChapterProgress(1000, 101)
	if err != nil {
		t.Fatalf("get progress: %v", err)
	}
	if stored.KillBossCount != 1 || stored.KillEnemyCount != 2 || stored.TakeBoxCount != 3 {
		t.Fatalf("unexpected stored counts")
	}

	progress.KillEnemyCount = 5
	progress.Progress = 100
	progress.PassCount = 1
	if err := UpsertChapterProgress(&progress); err != nil {
		t.Fatalf("upsert progress update: %v", err)
	}
	stored, err = GetChapterProgress(1000, 101)
	if err != nil {
		t.Fatalf("get progress after update: %v", err)
	}
	if stored.KillEnemyCount != 5 || stored.Progress != 100 || stored.PassCount != 1 {
		t.Fatalf("unexpected updated progress values")
	}
}

func TestDeleteChapterProgress(t *testing.T) {
	initBattleSessionTestDB(t)
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `DELETE FROM chapter_progress`); err != nil {
		t.Fatalf("clear chapter progress: %v", err)
	}
	progress := ChapterProgress{
		CommanderID: 2000,
		ChapterID:   202,
	}
	if err := UpsertChapterProgress(&progress); err != nil {
		t.Fatalf("upsert progress: %v", err)
	}
	if err := DeleteChapterProgress(2000, 202); err != nil {
		t.Fatalf("delete progress: %v", err)
	}
	_, err := GetChapterProgress(2000, 202)
	if !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("expected record not found after delete, got %v", err)
	}
}

func TestListChapterProgressOrdersByChapter(t *testing.T) {
	initBattleSessionTestDB(t)
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `DELETE FROM chapter_progress`); err != nil {
		t.Fatalf("clear chapter progress: %v", err)
	}
	entries := []ChapterProgress{
		{CommanderID: 3000, ChapterID: 3, Progress: 1},
		{CommanderID: 3000, ChapterID: 1, Progress: 2},
		{CommanderID: 3000, ChapterID: 2, Progress: 3},
	}
	for i := range entries {
		if err := UpsertChapterProgress(&entries[i]); err != nil {
			t.Fatalf("upsert progress: %v", err)
		}
	}
	list, err := ListChapterProgress(3000)
	if err != nil {
		t.Fatalf("list progress: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(list))
	}
	if list[0].ChapterID != 1 || list[1].ChapterID != 2 || list[2].ChapterID != 3 {
		t.Fatalf("unexpected order: %+v", list)
	}
}

func TestListChapterProgressPageUsesDBPaginationAndTotal(t *testing.T) {
	initBattleSessionTestDB(t)
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `DELETE FROM chapter_progress`); err != nil {
		t.Fatalf("clear chapter progress: %v", err)
	}
	entries := []ChapterProgress{
		{CommanderID: 3100, ChapterID: 1, Progress: 1},
		{CommanderID: 3100, ChapterID: 2, Progress: 2},
		{CommanderID: 3100, ChapterID: 3, Progress: 3},
	}
	for i := range entries {
		if err := UpsertChapterProgress(&entries[i]); err != nil {
			t.Fatalf("upsert progress: %v", err)
		}
	}

	result, err := ListChapterProgressPage(3100, 1, 1)
	if err != nil {
		t.Fatalf("list progress page: %v", err)
	}
	if result.Total != 3 {
		t.Fatalf("expected total 3, got %d", result.Total)
	}
	if len(result.Progress) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result.Progress))
	}
	if result.Progress[0].ChapterID != 2 {
		t.Fatalf("expected chapter_id 2, got %d", result.Progress[0].ChapterID)
	}
}

func TestSearchChapterProgressFiltersAndSortsByUpdatedAtDesc(t *testing.T) {
	initBattleSessionTestDB(t)
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `DELETE FROM chapter_progress`); err != nil {
		t.Fatalf("clear chapter progress: %v", err)
	}
	entries := []ChapterProgress{
		{CommanderID: 3200, ChapterID: 11, Progress: 10},
		{CommanderID: 3200, ChapterID: 12, Progress: 20},
		{CommanderID: 3200, ChapterID: 13, Progress: 30},
	}
	for i := range entries {
		if err := UpsertChapterProgress(&entries[i]); err != nil {
			t.Fatalf("upsert progress: %v", err)
		}
	}

	execUpdates := []struct {
		chapterID uint32
		updatedAt uint32
	}{
		{chapterID: 11, updatedAt: 100},
		{chapterID: 12, updatedAt: 300},
		{chapterID: 13, updatedAt: 200},
	}
	for _, update := range execUpdates {
		if _, err := db.DefaultStore.Pool.Exec(context.Background(), `
UPDATE chapter_progress
SET updated_at = $3
WHERE commander_id = $1 AND chapter_id = $2
`, int64(3200), int64(update.chapterID), int64(update.updatedAt)); err != nil {
			t.Fatalf("update updated_at: %v", err)
		}
	}

	updatedSince := uint32(150)
	result, err := SearchChapterProgress(3200, nil, &updatedSince, 0, 0)
	if err != nil {
		t.Fatalf("search chapter progress: %v", err)
	}
	if result.Total != 2 {
		t.Fatalf("expected total 2, got %d", result.Total)
	}
	if len(result.Progress) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(result.Progress))
	}
	if result.Progress[0].ChapterID != 12 || result.Progress[1].ChapterID != 13 {
		t.Fatalf("unexpected search order: %+v", result.Progress)
	}

	chapterID := uint32(13)
	filtered, err := SearchChapterProgress(3200, &chapterID, nil, 0, 0)
	if err != nil {
		t.Fatalf("search chapter progress by chapter_id: %v", err)
	}
	if filtered.Total != 1 || len(filtered.Progress) != 1 || filtered.Progress[0].ChapterID != 13 {
		t.Fatalf("unexpected chapter_id filter result: %+v", filtered)
	}
}
