package orm

import "gorm.io/gorm"

func UpdateCommanderRandomFlagShipEnabled(db *gorm.DB, commanderID uint32, enabled bool) error {
	return db.Model(&Commander{}).
		Where("commander_id = ?", commanderID).
		Update("random_flag_ship_enabled", enabled).Error
}
