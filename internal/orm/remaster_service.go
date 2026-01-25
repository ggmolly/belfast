package orm

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetOrCreateRemasterState(db *gorm.DB, commanderID uint32) (*RemasterState, error) {
	var state RemasterState
	if err := db.Where("commander_id = ?", commanderID).First(&state).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		state = RemasterState{
			CommanderID:      commanderID,
			LastDailyResetAt: time.Unix(0, 0),
		}
		if err := db.Create(&state).Error; err != nil {
			return nil, err
		}
	}
	return &state, nil
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

func ListRemasterProgress(db *gorm.DB, commanderID uint32) ([]RemasterProgress, error) {
	var progress []RemasterProgress
	if err := db.Where("commander_id = ?", commanderID).Order("chapter_id asc, pos asc").Find(&progress).Error; err != nil {
		return nil, err
	}
	return progress, nil
}

func GetRemasterProgress(db *gorm.DB, commanderID uint32, chapterID uint32, pos uint32) (*RemasterProgress, error) {
	var entry RemasterProgress
	if err := db.Where("commander_id = ? AND chapter_id = ? AND pos = ?", commanderID, chapterID, pos).First(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func UpsertRemasterProgress(db *gorm.DB, entry *RemasterProgress) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "commander_id"}, {Name: "chapter_id"}, {Name: "pos"}},
		DoUpdates: clause.AssignmentColumns([]string{"count", "received", "updated_at"}),
	}).Create(entry).Error
}

func DeleteRemasterProgress(db *gorm.DB, commanderID uint32, chapterID uint32, pos uint32) error {
	return db.Where("commander_id = ? AND chapter_id = ? AND pos = ?", commanderID, chapterID, pos).Delete(&RemasterProgress{}).Error
}

func startOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}
