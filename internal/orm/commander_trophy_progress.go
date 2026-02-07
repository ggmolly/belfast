package orm

import (
	"errors"

	"gorm.io/gorm"
)

// CommanderTrophyProgress tracks a commander's trophy/medal progress and claim timestamp.
// Timestamp == 0 means unclaimed.
type CommanderTrophyProgress struct {
	CommanderID uint32 `gorm:"primaryKey;autoIncrement:false"`
	TrophyID    uint32 `gorm:"primaryKey;autoIncrement:false"`
	Progress    uint32 `gorm:"not null;default:0"`
	Timestamp   uint32 `gorm:"not null;default:0"`
}

func GetCommanderTrophyProgress(db *gorm.DB, commanderID uint32, trophyID uint32) (*CommanderTrophyProgress, error) {
	var row CommanderTrophyProgress
	if err := db.Where("commander_id = ? AND trophy_id = ?", commanderID, trophyID).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func GetOrCreateCommanderTrophyProgress(db *gorm.DB, commanderID uint32, trophyID uint32, progress uint32) (*CommanderTrophyProgress, bool, error) {
	row, err := GetCommanderTrophyProgress(db, commanderID, trophyID)
	if err == nil {
		return row, false, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}
	row = &CommanderTrophyProgress{
		CommanderID: commanderID,
		TrophyID:    trophyID,
		Progress:    progress,
		Timestamp:   0,
	}
	if err := db.Create(row).Error; err != nil {
		return nil, false, err
	}
	return row, true, nil
}

func UpdateCommanderTrophyProgress(db *gorm.DB, row *CommanderTrophyProgress) error {
	return db.Save(row).Error
}

func ClaimCommanderTrophyProgress(db *gorm.DB, commanderID uint32, trophyID uint32, timestamp uint32) error {
	return db.Model(&CommanderTrophyProgress{}).
		Where("commander_id = ? AND trophy_id = ?", commanderID, trophyID).
		Update("timestamp", timestamp).Error
}
