package orm

import (
	"context"
	"time"

	"github.com/ggmolly/belfast/internal/db"
)

type ChapterState struct {
	CommanderID uint32 `gorm:"primary_key"`
	ChapterID   uint32 `gorm:"not_null;index"`
	State       []byte `gorm:"not_null"`
	UpdatedAt   uint32 `gorm:"not_null"`
}

func GetChapterState(commanderID uint32) (*ChapterState, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT commander_id, chapter_id, state, updated_at
FROM chapter_states
WHERE commander_id = $1
`, int64(commanderID))
	var state ChapterState
	err := row.Scan(&state.CommanderID, &state.ChapterID, &state.State, &state.UpdatedAt)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	// expire per-commander chapter state after 24h
	now := uint32(time.Now().Unix())
	if state.UpdatedAt != 0 && now-state.UpdatedAt > 60*60*24 {
		progress, err := GetChapterProgress(commanderID, state.ChapterID)
		if err == nil && progress.Progress >= 100 {
			return &state, nil
		}
		if _, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM chapter_states WHERE commander_id = $1`, int64(commanderID)); err != nil {
			return nil, err
		}
		return nil, db.ErrNotFound
	}
	return &state, nil
}

func UpsertChapterState(state *ChapterState) error {
	ctx := context.Background()
	state.UpdatedAt = uint32(time.Now().Unix())
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO chapter_states (commander_id, chapter_id, state, updated_at)
VALUES ($1, $2, $3, $4)
ON CONFLICT (commander_id)
DO UPDATE SET
  chapter_id = EXCLUDED.chapter_id,
  state = EXCLUDED.state,
  updated_at = EXCLUDED.updated_at
`, int64(state.CommanderID), int64(state.ChapterID), state.State, int64(state.UpdatedAt))
	return err
}

func ListChapterStates(commanderID uint32) ([]ChapterState, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT commander_id, chapter_id, state, updated_at
FROM chapter_states
WHERE commander_id = $1
ORDER BY updated_at DESC
`, int64(commanderID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	states := make([]ChapterState, 0)
	for rows.Next() {
		var state ChapterState
		if err := rows.Scan(&state.CommanderID, &state.ChapterID, &state.State, &state.UpdatedAt); err != nil {
			return nil, err
		}
		states = append(states, state)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return states, nil
}

func DeleteChapterState(commanderID uint32) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM chapter_states WHERE commander_id = $1`, int64(commanderID))
	return err
}
