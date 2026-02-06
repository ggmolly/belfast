package orm

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OwnedShipShadowSkin struct {
	CommanderID uint32 `gorm:"primaryKey;autoIncrement:false"`
	ShipID      uint32 `gorm:"primaryKey;autoIncrement:false"`
	ShadowID    uint32 `gorm:"primaryKey;autoIncrement:false"`
	SkinID      uint32 `gorm:"not_null"`
}

func UpsertOwnedShipShadowSkin(tx *gorm.DB, commanderID uint32, shipID uint32, shadowID uint32, skinID uint32) error {
	entry := OwnedShipShadowSkin{
		CommanderID: commanderID,
		ShipID:      shipID,
		ShadowID:    shadowID,
		SkinID:      skinID,
	}
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "commander_id"}, {Name: "ship_id"}, {Name: "shadow_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"skin_id"}),
	}).Create(&entry).Error
}

func ListOwnedShipShadowSkins(commanderID uint32, shipIDs []uint32) (map[uint32][]OwnedShipShadowSkin, error) {
	var entries []OwnedShipShadowSkin
	query := GormDB.Where("commander_id = ?", commanderID)
	if len(shipIDs) > 0 {
		query = query.Where("ship_id IN ?", shipIDs)
	}
	if err := query.Order("ship_id asc").Order("shadow_id asc").Find(&entries).Error; err != nil {
		return nil, err
	}
	result := make(map[uint32][]OwnedShipShadowSkin)
	for _, entry := range entries {
		result[entry.ShipID] = append(result[entry.ShipID], entry)
	}
	return result, nil
}
