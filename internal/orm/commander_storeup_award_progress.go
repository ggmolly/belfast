package orm

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CommanderStoreupAwardProgress tracks which collection (storeup) award tiers a commander has claimed.
// LastAwardIndex is a 1-based index into storeup_data_template.award_display/level.
type CommanderStoreupAwardProgress struct {
	CommanderID    uint32 `gorm:"primaryKey;autoIncrement:false"`
	StoreupID      uint32 `gorm:"primaryKey;autoIncrement:false"`
	LastAwardIndex uint32 `gorm:"not null;default:0"`
}

func GetCommanderStoreupAwardProgress(db *gorm.DB, commanderID uint32, storeupID uint32) (*CommanderStoreupAwardProgress, error) {
	var row CommanderStoreupAwardProgress
	if err := db.Where("commander_id = ? AND storeup_id = ?", commanderID, storeupID).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func GetLastCommanderStoreupAwardIndex(db *gorm.DB, commanderID uint32, storeupID uint32) (uint32, error) {
	row, err := GetCommanderStoreupAwardProgress(db, commanderID, storeupID)
	if err == nil {
		return row.LastAwardIndex, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	return 0, err
}

func ListCommanderStoreupAwardProgress(db *gorm.DB, commanderID uint32) ([]CommanderStoreupAwardProgress, error) {
	var rows []CommanderStoreupAwardProgress
	if err := db.Where("commander_id = ?", commanderID).Order("storeup_id asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func SetCommanderStoreupAwardIndexTx(tx *gorm.DB, commanderID uint32, storeupID uint32, lastAwardIndex uint32) error {
	row := CommanderStoreupAwardProgress{CommanderID: commanderID, StoreupID: storeupID, LastAwardIndex: lastAwardIndex}
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "commander_id"}, {Name: "storeup_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"last_award_index"}),
	}).Create(&row).Error
}
