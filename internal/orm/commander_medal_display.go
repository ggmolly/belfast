package orm

import (
	"gorm.io/gorm"
)

// CommanderMedalDisplay persists a commander's ordered medal/trophy display list.
// Ordering is stable via Position (0-based).
type CommanderMedalDisplay struct {
	CommanderID uint32 `gorm:"primaryKey;autoIncrement:false"`
	Position    uint32 `gorm:"primaryKey;autoIncrement:false"`
	MedalID     uint32 `gorm:"not null"`
}

func ListCommanderMedalDisplay(commanderID uint32) ([]uint32, error) {
	var entries []CommanderMedalDisplay
	if err := GormDB.
		Where("commander_id = ?", commanderID).
		Order("position ASC").
		Find(&entries).Error; err != nil {
		return nil, err
	}
	medalIDs := make([]uint32, 0, len(entries))
	for _, entry := range entries {
		medalIDs = append(medalIDs, entry.MedalID)
	}
	return medalIDs, nil
}

func SetCommanderMedalDisplay(db *gorm.DB, commanderID uint32, medalIDs []uint32) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("commander_id = ?", commanderID).Delete(&CommanderMedalDisplay{}).Error; err != nil {
			return err
		}
		if len(medalIDs) == 0 {
			return nil
		}
		rows := make([]CommanderMedalDisplay, 0, len(medalIDs))
		for i, medalID := range medalIDs {
			rows = append(rows, CommanderMedalDisplay{
				CommanderID: commanderID,
				Position:    uint32(i),
				MedalID:     medalID,
			})
		}
		return tx.Create(&rows).Error
	})
}
