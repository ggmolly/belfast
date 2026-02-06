package orm

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ChapterDrop tracks unique ship drops for a commander in a given chapter.
// Rows are inserted as ships are obtained; duplicates are ignored.
type ChapterDrop struct {
	CommanderID uint32 `gorm:"primaryKey;autoIncrement:false"`
	ChapterID   uint32 `gorm:"primaryKey;autoIncrement:false"`
	ShipID      uint32 `gorm:"primaryKey;autoIncrement:false"`
}

func GetChapterDrops(db *gorm.DB, commanderID uint32, chapterID uint32) ([]ChapterDrop, error) {
	var drops []ChapterDrop
	if err := db.Where("commander_id = ? AND chapter_id = ?", commanderID, chapterID).Order("ship_id asc").Find(&drops).Error; err != nil {
		return nil, err
	}
	return drops, nil
}

func AddChapterDrop(db *gorm.DB, drop *ChapterDrop) error {
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(drop).Error
}
