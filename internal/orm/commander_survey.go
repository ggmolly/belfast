package orm

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CommanderSurvey struct {
	CommanderID uint32    `gorm:"primaryKey;autoIncrement:false"`
	SurveyID    uint32    `gorm:"primaryKey;autoIncrement:false"`
	CompletedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
}

func IsCommanderSurveyCompleted(commanderID uint32, surveyID uint32) (bool, error) {
	var count int64
	if err := GormDB.Model(&CommanderSurvey{}).Where("commander_id = ? AND survey_id = ?", commanderID, surveyID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func SetCommanderSurveyCompleted(db *gorm.DB, commanderID uint32, surveyID uint32, completedAt time.Time) error {
	entry := CommanderSurvey{CommanderID: commanderID, SurveyID: surveyID, CompletedAt: completedAt}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "commander_id"}, {Name: "survey_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"completed_at"}),
	}).Create(&entry).Error
}
