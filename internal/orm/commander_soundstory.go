package orm

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CommanderSoundStory struct {
	CommanderID  uint32    `gorm:"primaryKey;autoIncrement:false"`
	SoundStoryID uint32    `gorm:"primaryKey;autoIncrement:false"`
	CreatedAt    time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
}

func ListCommanderSoundStoryIDs(commanderID uint32) ([]uint32, error) {
	var entries []CommanderSoundStory
	if err := GormDB.Where("commander_id = ?", commanderID).Order("sound_story_id asc").Find(&entries).Error; err != nil {
		return nil, err
	}
	ids := make([]uint32, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.SoundStoryID)
	}
	return ids, nil
}

func IsCommanderSoundStoryUnlockedTx(tx *gorm.DB, commanderID uint32, soundStoryID uint32) (bool, error) {
	var entry CommanderSoundStory
	result := tx.Where("commander_id = ? AND sound_story_id = ?", commanderID, soundStoryID).Limit(1).Find(&entry)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func UnlockCommanderSoundStoryTx(tx *gorm.DB, commanderID uint32, soundStoryID uint32) error {
	entry := CommanderSoundStory{CommanderID: commanderID, SoundStoryID: soundStoryID}
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&entry).Error
}
