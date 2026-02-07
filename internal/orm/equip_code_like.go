package orm

import "time"

type EquipCodeLike struct {
	ID uint64 `gorm:"primaryKey"`

	CommanderID uint32 `gorm:"not_null;index;uniqueIndex:idx_equip_code_like_dedupe"`
	ShipGroupID uint32 `gorm:"not_null;uniqueIndex:idx_equip_code_like_dedupe"`
	ShareID     uint32 `gorm:"not_null;index;uniqueIndex:idx_equip_code_like_dedupe"`
	LikeDay     uint32 `gorm:"not_null;uniqueIndex:idx_equip_code_like_dedupe"`
	CreatedAt   time.Time
}
