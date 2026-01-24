package orm

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CommanderLivingAreaCover struct {
	CommanderID uint32    `gorm:"primaryKey;autoIncrement:false"`
	CoverID     uint32    `gorm:"primaryKey;autoIncrement:false"`
	UnlockedAt  time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
	IsNew       bool      `gorm:"default:false;not_null"`
}

func ListCommanderLivingAreaCovers(commanderID uint32) ([]CommanderLivingAreaCover, error) {
	var entries []CommanderLivingAreaCover
	if err := GormDB.Where("commander_id = ?", commanderID).Order("cover_id asc").Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

func CommanderHasLivingAreaCover(commanderID uint32, coverID uint32) (bool, error) {
	var entry CommanderLivingAreaCover
	if err := GormDB.Where("commander_id = ? AND cover_id = ?", commanderID, coverID).First(&entry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func UpsertCommanderLivingAreaCover(db *gorm.DB, entry CommanderLivingAreaCover) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "commander_id"}, {Name: "cover_id"}},
		UpdateAll: true,
	}).Create(&entry).Error
}
