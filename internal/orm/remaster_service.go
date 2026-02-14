package orm

import (
	"context"
	"time"

	"github.com/ggmolly/belfast/internal/db"
)

func GetOrCreateRemasterState(commanderID uint32) (*RemasterState, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT commander_id, ticket_count, active_chapter_id, daily_count, last_daily_reset_at, created_at, updated_at
FROM remaster_states
WHERE commander_id = $1
`, int64(commanderID))
	var state RemasterState
	err := row.Scan(&state.CommanderID, &state.TicketCount, &state.ActiveChapterID, &state.DailyCount, &state.LastDailyResetAt, &state.CreatedAt, &state.UpdatedAt)
	err = db.MapNotFound(err)
	if err == nil {
		return &state, nil
	}
	if !db.IsNotFound(err) {
		return nil, err
	}
	state = RemasterState{CommanderID: commanderID, LastDailyResetAt: time.Unix(0, 0)}
	if err := SaveRemasterState(&state); err != nil {
		return nil, err
	}
	return &state, nil
}

func SaveRemasterState(state *RemasterState) error {
	ctx := context.Background()
	now := time.Now().UTC()
	if state.LastDailyResetAt.IsZero() {
		state.LastDailyResetAt = time.Unix(0, 0)
	}
	if state.CreatedAt.IsZero() {
		state.CreatedAt = now
	}
	state.UpdatedAt = now
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO remaster_states (
  commander_id,
  ticket_count,
  active_chapter_id,
  daily_count,
  last_daily_reset_at,
  created_at,
  updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
ON CONFLICT (commander_id)
DO UPDATE SET
  ticket_count = EXCLUDED.ticket_count,
  active_chapter_id = EXCLUDED.active_chapter_id,
  daily_count = EXCLUDED.daily_count,
  last_daily_reset_at = EXCLUDED.last_daily_reset_at,
  updated_at = EXCLUDED.updated_at
`, int64(state.CommanderID), int64(state.TicketCount), int64(state.ActiveChapterID), int64(state.DailyCount), state.LastDailyResetAt, state.CreatedAt, state.UpdatedAt)
	return err
}

func ApplyRemasterDailyReset(state *RemasterState, now time.Time) bool {
	resetAt := startOfDay(now)
	if state.LastDailyResetAt.Before(resetAt) {
		state.DailyCount = 0
		state.LastDailyResetAt = resetAt
		return true
	}
	return false
}

func ListRemasterProgress(commanderID uint32) ([]RemasterProgress, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT id, commander_id, chapter_id, pos, count, received, created_at, updated_at
FROM remaster_progresses
WHERE commander_id = $1
ORDER BY chapter_id ASC, pos ASC
`, int64(commanderID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]RemasterProgress, 0)
	for rows.Next() {
		var entry RemasterProgress
		if err := rows.Scan(&entry.ID, &entry.CommanderID, &entry.ChapterID, &entry.Pos, &entry.Count, &entry.Received, &entry.CreatedAt, &entry.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func GetRemasterProgress(commanderID uint32, chapterID uint32, pos uint32) (*RemasterProgress, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT id, commander_id, chapter_id, pos, count, received, created_at, updated_at
FROM remaster_progresses
WHERE commander_id = $1 AND chapter_id = $2 AND pos = $3
`, int64(commanderID), int64(chapterID), int64(pos))
	var entry RemasterProgress
	err := row.Scan(&entry.ID, &entry.CommanderID, &entry.ChapterID, &entry.Pos, &entry.Count, &entry.Received, &entry.CreatedAt, &entry.UpdatedAt)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func UpsertRemasterProgress(entry *RemasterProgress) error {
	ctx := context.Background()
	now := time.Now().UTC()
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = now
	}
	entry.UpdatedAt = now
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO remaster_progresses (
  commander_id,
  chapter_id,
  pos,
  count,
  received,
  created_at,
  updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
ON CONFLICT (commander_id, chapter_id, pos)
DO UPDATE SET
  count = EXCLUDED.count,
  received = EXCLUDED.received,
  updated_at = EXCLUDED.updated_at
`, int64(entry.CommanderID), int64(entry.ChapterID), int64(entry.Pos), int64(entry.Count), entry.Received, entry.CreatedAt, entry.UpdatedAt)
	return err
}

func DeleteRemasterProgress(commanderID uint32, chapterID uint32, pos uint32) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM remaster_progresses
WHERE commander_id = $1 AND chapter_id = $2 AND pos = $3
`, int64(commanderID), int64(chapterID), int64(pos))
	return err
}

func startOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}
