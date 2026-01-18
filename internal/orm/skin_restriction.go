package orm

// GlobalSkinRestriction defines global hide/show restrictions for skins.
type GlobalSkinRestriction struct {
	SkinID uint32 `gorm:"primaryKey"`
	Type   uint32 `gorm:"not_null"`
}

// GlobalSkinRestrictionWindow defines time-based overrides for skin shop availability.
type GlobalSkinRestrictionWindow struct {
	ID        uint32 `gorm:"primaryKey"`
	SkinID    uint32 `gorm:"not_null"`
	Type      uint32 `gorm:"not_null"`
	StartTime uint32 `gorm:"not_null"`
	StopTime  uint32 `gorm:"not_null"`
}
