package orm

import "time"

// EquipCodeShare tracks per-commander sharing activity for equipment codes.
// It is used for enforcing daily share limits.
type EquipCodeShare struct {
	ID uint64 `gorm:"primaryKey"`

	CommanderID uint32 `gorm:"not_null;index;uniqueIndex:idx_equip_code_share_dedupe"`
	ShipGroupID uint32 `gorm:"not_null;uniqueIndex:idx_equip_code_share_dedupe"`
	ShareDay    uint32 `gorm:"not_null;uniqueIndex:idx_equip_code_share_dedupe"`
	CreatedAt   time.Time
}
