package orm

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CommanderAttire struct {
	CommanderID uint32     `gorm:"primaryKey;autoIncrement:false"`
	Type        uint32     `gorm:"primaryKey;autoIncrement:false"`
	AttireID    uint32     `gorm:"primaryKey;autoIncrement:false"`
	ExpiresAt   *time.Time `gorm:"type:timestamp"`
	IsNew       bool       `gorm:"default:false;not_null"`
}

func ListCommanderAttires(commanderID uint32) ([]CommanderAttire, error) {
	var entries []CommanderAttire
	if err := GormDB.Where("commander_id = ?", commanderID).Order("type asc, attire_id asc").Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

func ListCommanderAttiresByType(commanderID uint32, attireType uint32) ([]CommanderAttire, error) {
	var entries []CommanderAttire
	if err := GormDB.Where("commander_id = ? AND type = ?", commanderID, attireType).Order("attire_id asc").Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

func CommanderHasAttire(commanderID uint32, attireType uint32, attireID uint32, now time.Time) (bool, error) {
	var entry CommanderAttire
	if err := GormDB.Where("commander_id = ? AND type = ? AND attire_id = ?", commanderID, attireType, attireID).
		First(&entry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	if entry.ExpiresAt != nil && entry.ExpiresAt.Before(now) {
		return false, nil
	}
	return true, nil
}

func UpsertCommanderAttire(db *gorm.DB, entry CommanderAttire) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "commander_id"}, {Name: "type"}, {Name: "attire_id"}},
		UpdateAll: true,
	}).Create(&entry).Error
}
