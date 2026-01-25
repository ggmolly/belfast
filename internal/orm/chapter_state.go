package orm

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ChapterState struct {
	CommanderID uint32 `gorm:"primary_key"`
	ChapterID   uint32 `gorm:"not_null;index"`
	State       []byte `gorm:"type:blob;not_null"`
	UpdatedAt   uint32 `gorm:"not_null"`
}

func GetChapterState(db *gorm.DB, commanderID uint32) (*ChapterState, error) {
	var state ChapterState
	if err := db.Where("commander_id = ?", commanderID).First(&state).Error; err != nil {
		return nil, err
	}
	return &state, nil
}

func UpsertChapterState(db *gorm.DB, state *ChapterState) error {
	state.UpdatedAt = uint32(time.Now().Unix())
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "commander_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"chapter_id", "state", "updated_at"}),
	}).Create(state).Error
}

func DeleteChapterState(db *gorm.DB, commanderID uint32) error {
	return db.Where("commander_id = ?", commanderID).Delete(&ChapterState{}).Error
}
