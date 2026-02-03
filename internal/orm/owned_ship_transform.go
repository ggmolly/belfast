package orm

import "gorm.io/gorm"

type OwnedShipTransform struct {
	OwnerID     uint32 `gorm:"primaryKey;autoIncrement:false"`
	ShipID      uint32 `gorm:"primaryKey;autoIncrement:false"`
	TransformID uint32 `gorm:"primaryKey;autoIncrement:false"`
	Level       uint32 `gorm:"not_null"`
}

func ListOwnedShipTransforms(db *gorm.DB, ownerID uint32, shipID uint32) ([]OwnedShipTransform, error) {
	var entries []OwnedShipTransform
	if err := db.Where("owner_id = ? AND ship_id = ?", ownerID, shipID).Order("transform_id asc").Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

func UpsertOwnedShipTransformTx(tx *gorm.DB, entry *OwnedShipTransform) error {
	return tx.Save(entry).Error
}

func DeleteOwnedShipTransformsTx(tx *gorm.DB, ownerID uint32, shipID uint32, transformIDs []uint32) error {
	if len(transformIDs) == 0 {
		return nil
	}
	return tx.Where("owner_id = ? AND ship_id = ? AND transform_id IN ?", ownerID, shipID, transformIDs).Delete(&OwnedShipTransform{}).Error
}
