package orm

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CommanderCommonFlag struct {
	CommanderID uint32    `gorm:"primaryKey;autoIncrement:false"`
	FlagID      uint32    `gorm:"primaryKey;autoIncrement:false"`
	CreatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
}

func ListCommanderCommonFlags(commanderID uint32) ([]uint32, error) {
	var entries []CommanderCommonFlag
	if err := GormDB.Where("commander_id = ?", commanderID).Order("flag_id asc").Find(&entries).Error; err != nil {
		return nil, err
	}
	flags := make([]uint32, 0, len(entries))
	for _, entry := range entries {
		flags = append(flags, entry.FlagID)
	}
	return flags, nil
}

func SetCommanderCommonFlag(db *gorm.DB, commanderID uint32, flagID uint32) error {
	entry := CommanderCommonFlag{CommanderID: commanderID, FlagID: flagID}
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&entry).Error
}

func ClearCommanderCommonFlag(db *gorm.DB, commanderID uint32, flagID uint32) error {
	return db.Where("commander_id = ? AND flag_id = ?", commanderID, flagID).Delete(&CommanderCommonFlag{}).Error
}
