package orm

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SurveyState struct {
	CommanderID uint32    `gorm:"primaryKey;autoIncrement:false"`
	SurveyID    uint32    `gorm:"not_null;default:0"`
	CompletedAt time.Time `gorm:"type:timestamp;default:'1970-01-01 00:00:00';not_null"`
	CreatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
	UpdatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
}

func GetSurveyState(db *gorm.DB, commanderID uint32) (*SurveyState, error) {
	var state SurveyState
	if err := db.Where("commander_id = ?", commanderID).First(&state).Error; err != nil {
		return nil, err
	}
	return &state, nil
}

func UpsertSurveyState(db *gorm.DB, state *SurveyState) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "commander_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"survey_id", "completed_at", "updated_at"}),
	}).Create(state).Error
}
