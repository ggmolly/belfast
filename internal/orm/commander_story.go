package orm

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CommanderStory struct {
	CommanderID uint32    `gorm:"primaryKey;autoIncrement:false"`
	StoryID     uint32    `gorm:"primaryKey;autoIncrement:false"`
	CreatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
}

func ListCommanderStoryIDs(commanderID uint32) ([]uint32, error) {
	var entries []CommanderStory
	if err := GormDB.Where("commander_id = ?", commanderID).Order("story_id asc").Find(&entries).Error; err != nil {
		return nil, err
	}
	ids := make([]uint32, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.StoryID)
	}
	return ids, nil
}

func AddCommanderStory(db *gorm.DB, commanderID uint32, storyID uint32) error {
	entry := CommanderStory{CommanderID: commanderID, StoryID: storyID}
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&entry).Error
}
