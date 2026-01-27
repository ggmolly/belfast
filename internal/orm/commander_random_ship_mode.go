package orm

import (
	"github.com/ggmolly/belfast/internal/consts"
	"gorm.io/gorm"
)

func UpdateCommanderRandomShipMode(db *gorm.DB, commanderID uint32, mode uint32) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Commander{}).
			Where("commander_id = ?", commanderID).
			Update("random_ship_mode", mode).Error; err != nil {
			return err
		}
		if mode == 2 {
			return SetCommanderCommonFlag(tx, commanderID, consts.RandomFlagShipMode)
		}
		return ClearCommanderCommonFlag(tx, commanderID, consts.RandomFlagShipMode)
	})
}
