package orm

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SubmarineExpeditionState struct {
	CommanderID        uint32 `gorm:"primary_key"`
	LastRefreshTime    uint32 `gorm:"not_null"`
	WeeklyRefreshCount uint32 `gorm:"not_null"`
	ActiveChapterID    uint32 `gorm:"not_null"`
	OverallProgress    uint32 `gorm:"not_null"`
}

func GetSubmarineState(db *gorm.DB, commanderID uint32) (*SubmarineExpeditionState, error) {
	var state SubmarineExpeditionState
	if err := db.Where("commander_id = ?", commanderID).First(&state).Error; err != nil {
		return nil, err
	}
	return &state, nil
}

func UpsertSubmarineState(db *gorm.DB, state *SubmarineExpeditionState) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "commander_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"last_refresh_time", "weekly_refresh_count", "active_chapter_id", "overall_progress"}),
	}).Create(state).Error
}

func ResetWeeklyRefresh(db *gorm.DB, commanderID uint32, refreshAt uint32) error {
	return db.Model(&SubmarineExpeditionState{}).
		Where("commander_id = ?", commanderID).
		Updates(map[string]any{
			"weekly_refresh_count": uint32(0),
			"last_refresh_time":    refreshAt,
		}).Error
}
