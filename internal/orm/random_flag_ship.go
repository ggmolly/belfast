package orm

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RandomFlagShip struct {
	CommanderID uint32 `gorm:"primaryKey;autoIncrement:false"`
	ShipID      uint32 `gorm:"primaryKey;autoIncrement:false"`
	PhantomID   uint32 `gorm:"primaryKey;autoIncrement:false"`
	Enabled     bool   `gorm:"default:true;not_null"`
}

type RandomFlagShipUpdate struct {
	ShipID    uint32
	PhantomID uint32
	Flag      uint32
}

func ApplyRandomFlagShipUpdates(tx *gorm.DB, commanderID uint32, updates []RandomFlagShipUpdate) error {
	for _, update := range updates {
		if update.Flag == 0 {
			if err := tx.Where("commander_id = ? AND ship_id = ? AND phantom_id = ?", commanderID, update.ShipID, update.PhantomID).
				Delete(&RandomFlagShip{}).Error; err != nil {
				return err
			}
			continue
		}
		entry := RandomFlagShip{
			CommanderID: commanderID,
			ShipID:      update.ShipID,
			PhantomID:   update.PhantomID,
			Enabled:     true,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "commander_id"}, {Name: "ship_id"}, {Name: "phantom_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"enabled"}),
		}).Create(&entry).Error; err != nil {
			return err
		}
	}
	return nil
}

func ListRandomFlagShipPhantoms(commanderID uint32, shipIDs []uint32) (map[uint32][]uint32, error) {
	flags := make(map[uint32][]uint32)
	var entries []RandomFlagShip
	query := GormDB.Where("commander_id = ? AND enabled = ?", commanderID, true)
	if len(shipIDs) > 0 {
		query = query.Where("ship_id IN ?", shipIDs)
	}
	if err := query.Find(&entries).Error; err != nil {
		return nil, err
	}
	for _, entry := range entries {
		flags[entry.ShipID] = append(flags[entry.ShipID], entry.PhantomID)
	}
	return flags, nil
}
