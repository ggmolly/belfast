package orm

type RequisitionShip struct {
	ShipID uint32 `gorm:"primary_key" json:"ship_id"`
}

func ListRequisitionShipIDs() ([]uint32, error) {
	var entries []RequisitionShip
	if err := GormDB.Find(&entries).Error; err != nil {
		return nil, err
	}
	ids := make([]uint32, len(entries))
	for i, entry := range entries {
		ids[i] = entry.ShipID
	}
	return ids, nil
}

func GetRandomRequisitionShipByRarity(rarity uint32) (Ship, error) {
	var ship Ship
	err := GormDB.Model(&Ship{}).
		Joins("JOIN requisition_ships ON requisition_ships.ship_id = ships.template_id").
		Where("ships.rarity_id = ?", rarity).
		Order("RANDOM()").
		First(&ship).Error
	return ship, err
}
