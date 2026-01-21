package orm

type RequisitionShip struct {
	ShipID uint32 `gorm:"primary_key" json:"ship_id"`
}
