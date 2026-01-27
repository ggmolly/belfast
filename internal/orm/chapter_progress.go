package orm

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ChapterProgress struct {
	CommanderID      uint32 `gorm:"primaryKey;autoIncrement:false"`
	ChapterID        uint32 `gorm:"primaryKey;autoIncrement:false"`
	Progress         uint32 `gorm:"not_null"`
	KillBossCount    uint32 `gorm:"not_null"`
	KillEnemyCount   uint32 `gorm:"not_null"`
	TakeBoxCount     uint32 `gorm:"not_null"`
	DefeatCount      uint32 `gorm:"not_null"`
	TodayDefeatCount uint32 `gorm:"not_null"`
	PassCount        uint32 `gorm:"not_null"`
	UpdatedAt        uint32 `gorm:"not_null"`
}

func GetChapterProgress(db *gorm.DB, commanderID uint32, chapterID uint32) (*ChapterProgress, error) {
	var progress ChapterProgress
	if err := db.Where("commander_id = ? AND chapter_id = ?", commanderID, chapterID).First(&progress).Error; err != nil {
		return nil, err
	}
	return &progress, nil
}

func ListChapterProgress(db *gorm.DB, commanderID uint32) ([]ChapterProgress, error) {
	var progress []ChapterProgress
	if err := db.Where("commander_id = ?", commanderID).Order("chapter_id asc").Find(&progress).Error; err != nil {
		return nil, err
	}
	return progress, nil
}

func UpsertChapterProgress(db *gorm.DB, progress *ChapterProgress) error {
	progress.UpdatedAt = uint32(time.Now().Unix())
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "commander_id"}, {Name: "chapter_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"progress", "kill_boss_count", "kill_enemy_count", "take_box_count", "defeat_count", "today_defeat_count", "pass_count", "updated_at"}),
	}).Create(progress).Error
}

func DeleteChapterProgress(db *gorm.DB, commanderID uint32, chapterID uint32) error {
	return db.Where("commander_id = ? AND chapter_id = ?", commanderID, chapterID).Delete(&ChapterProgress{}).Error
}
