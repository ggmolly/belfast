package orm

import (
	"context"
	"time"

	"github.com/ggmolly/belfast/internal/db"
)

type SurveyState struct {
	CommanderID uint32    `gorm:"primaryKey;autoIncrement:false"`
	SurveyID    uint32    `gorm:"not_null;default:0"`
	CompletedAt time.Time `gorm:"type:timestamp;default:'1970-01-01 00:00:00';not_null"`
	CreatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
	UpdatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
}

func GetSurveyState(commanderID uint32) (*SurveyState, error) {
	ctx := context.Background()
	state := SurveyState{}
	err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT commander_id, survey_id, completed_at, created_at, updated_at
FROM survey_states
WHERE commander_id = $1
`, int64(commanderID)).Scan(&state.CommanderID, &state.SurveyID, &state.CompletedAt, &state.CreatedAt, &state.UpdatedAt)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func UpsertSurveyState(state *SurveyState) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO survey_states (commander_id, survey_id, completed_at, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
ON CONFLICT (commander_id)
DO UPDATE SET
  survey_id = EXCLUDED.survey_id,
  completed_at = EXCLUDED.completed_at,
  updated_at = NOW()
`, int64(state.CommanderID), int64(state.SurveyID), state.CompletedAt)
	return err
}
