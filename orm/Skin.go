package orm

import "time"

type Skin struct {
	ID        uint32 `gorm:"primary_key"`
	Name      string `gorm:"size:128;not_null"`
	ShipGroup int    `gorm:"not_null"`
}

type OwnedSkin struct {
	CommanderID uint32     `gorm:"primaryKey"`
	SkinID      uint32     `gorm:"primaryKey"`
	ExpiresAt   *time.Time `gorm:"not_null"`
}
