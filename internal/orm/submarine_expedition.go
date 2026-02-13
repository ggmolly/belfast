package orm

import (
	"context"

	"github.com/ggmolly/belfast/internal/db"
)

type SubmarineExpeditionState struct {
	CommanderID        uint32 `gorm:"primary_key"`
	LastRefreshTime    uint32 `gorm:"not_null"`
	WeeklyRefreshCount uint32 `gorm:"not_null"`
	ActiveChapterID    uint32 `gorm:"not_null"`
	OverallProgress    uint32 `gorm:"not_null"`
}

func GetSubmarineState(commanderID uint32) (*SubmarineExpeditionState, error) {
	ctx := context.Background()
	state := SubmarineExpeditionState{}
	err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT commander_id, last_refresh_time, weekly_refresh_count, active_chapter_id, overall_progress
FROM submarine_expedition_states
WHERE commander_id = $1
`, int64(commanderID)).Scan(&state.CommanderID, &state.LastRefreshTime, &state.WeeklyRefreshCount, &state.ActiveChapterID, &state.OverallProgress)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func UpsertSubmarineState(state *SubmarineExpeditionState) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO submarine_expedition_states (commander_id, last_refresh_time, weekly_refresh_count, active_chapter_id, overall_progress)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (commander_id)
DO UPDATE SET
  last_refresh_time = EXCLUDED.last_refresh_time,
  weekly_refresh_count = EXCLUDED.weekly_refresh_count,
  active_chapter_id = EXCLUDED.active_chapter_id,
  overall_progress = EXCLUDED.overall_progress
`, int64(state.CommanderID), int64(state.LastRefreshTime), int64(state.WeeklyRefreshCount), int64(state.ActiveChapterID), int64(state.OverallProgress))
	return err
}

func ResetWeeklyRefresh(commanderID uint32, refreshAt uint32) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE submarine_expedition_states
SET weekly_refresh_count = 0, last_refresh_time = $2
WHERE commander_id = $1
`, int64(commanderID), int64(refreshAt))
	return err
}
