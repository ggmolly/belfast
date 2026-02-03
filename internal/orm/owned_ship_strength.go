package orm

import "gorm.io/gorm"

type OwnedShipStrength struct {
	OwnerID    uint32 `gorm:"primaryKey;autoIncrement:false"`
	ShipID     uint32 `gorm:"primaryKey;autoIncrement:false"`
	StrengthID uint32 `gorm:"primaryKey;autoIncrement:false"`
	Exp        uint32 `gorm:"not_null"`
}

func ListOwnedShipStrengths(db *gorm.DB, ownerID uint32, shipID uint32) ([]OwnedShipStrength, error) {
	var entries []OwnedShipStrength
	if err := db.Where("owner_id = ? AND ship_id = ?", ownerID, shipID).Order("strength_id asc").Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

func UpsertOwnedShipStrengthTx(tx *gorm.DB, entry *OwnedShipStrength) error {
	return tx.Save(entry).Error
}
