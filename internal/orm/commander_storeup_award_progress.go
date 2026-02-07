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

// TryAdvanceCommanderStoreupAwardIndexTx atomically advances the storeup progress by exactly one tier.
//
// It returns (true, nil) only if the index was advanced from (awardIndex-1) -> awardIndex.
// This is used to prevent duplicate claims on concurrent requests.
func TryAdvanceCommanderStoreupAwardIndexTx(tx *gorm.DB, commanderID uint32, storeupID uint32, awardIndex uint32) (bool, error) {
	if awardIndex == 0 {
		return false, errors.New("award index must be > 0")
	}

	// Tier 1: allow insert, or update only if currently 0.
	if awardIndex == 1 {
		row := CommanderStoreupAwardProgress{CommanderID: commanderID, StoreupID: storeupID, LastAwardIndex: awardIndex}
		res := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "commander_id"}, {Name: "storeup_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"last_award_index"}),
			Where: clause.Where{Exprs: []clause.Expression{
				clause.Eq{Column: "last_award_index", Value: uint32(0)},
			}},
		}).Create(&row)
		if res.Error != nil {
			return false, res.Error
		}
		return res.RowsAffected == 1, nil
	}

	previousIndex := awardIndex - 1
	res := tx.Model(&CommanderStoreupAwardProgress{}).
		Where("commander_id = ? AND storeup_id = ? AND last_award_index = ?", commanderID, storeupID, previousIndex).
		Update("last_award_index", awardIndex)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected == 1, nil
}
