package orm

import (
	"context"
	"time"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

type CommanderSurvey struct {
	CommanderID uint32    `gorm:"primaryKey;autoIncrement:false"`
	SurveyID    uint32    `gorm:"primaryKey;autoIncrement:false"`
	CompletedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
}

func IsCommanderSurveyCompleted(commanderID uint32, surveyID uint32) (bool, error) {
	ctx := context.Background()
	exists, err := db.DefaultStore.Queries.HasCommanderSurvey(ctx, gen.HasCommanderSurveyParams{CommanderID: int64(commanderID), SurveyID: int64(surveyID)})
	return exists, err
}

func SetCommanderSurveyCompleted(commanderID uint32, surveyID uint32, completedAt time.Time) error {
	ctx := context.Background()
	return db.DefaultStore.Queries.UpsertCommanderSurvey(ctx, gen.UpsertCommanderSurveyParams{CommanderID: int64(commanderID), SurveyID: int64(surveyID), CompletedAt: pgTimestamptz(completedAt)})
}
