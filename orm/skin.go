package orm

import "time"

type Skin struct {
	ID        uint32 `gorm:"primary_key" json:"id"`
	Name      string `gorm:"size:128;not_null" json:"name"`
	ShipGroup int    `gorm:"not_null" json:"ship_group"`
}

type OwnedSkin struct {
	CommanderID uint32     `gorm:"primaryKey"`
	SkinID      uint32     `gorm:"primaryKey"`
	ExpiresAt   *time.Time `gorm:"not_null"`
}
