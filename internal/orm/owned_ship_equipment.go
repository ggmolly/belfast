package orm

import "gorm.io/gorm"

type OwnedShipEquipment struct {
	OwnerID uint32 `gorm:"primaryKey;autoIncrement:false" json:"owner_id"`
	ShipID  uint32 `gorm:"primaryKey;autoIncrement:false" json:"ship_id"`
	Pos     uint32 `gorm:"primaryKey;autoIncrement:false" json:"pos"`
	EquipID uint32 `gorm:"not_null" json:"equip_id"`
	SkinID  uint32 `gorm:"not_null" json:"skin_id"`

	Equipment Equipment `gorm:"foreignKey:EquipID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

func ListOwnedShipEquipment(db *gorm.DB, ownerID uint32, shipID uint32) ([]OwnedShipEquipment, error) {
	var entries []OwnedShipEquipment
	if err := db.Where("owner_id = ? AND ship_id = ?", ownerID, shipID).Order("pos asc").Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

func GetOwnedShipEquipment(db *gorm.DB, ownerID uint32, shipID uint32, pos uint32) (*OwnedShipEquipment, error) {
	var entry OwnedShipEquipment
	if err := db.Where("owner_id = ? AND ship_id = ? AND pos = ?", ownerID, shipID, pos).First(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func UpsertOwnedShipEquipmentTx(tx *gorm.DB, entry *OwnedShipEquipment) error {
	return tx.Save(entry).Error
}
