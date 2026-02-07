package orm

import "time"

type EquipCodeReport struct {
	ID uint64 `gorm:"primaryKey"`

	CommanderID uint32 `gorm:"not_null;index;uniqueIndex:idx_equip_code_report_dedupe"`
	ShareID     uint32 `gorm:"not_null;index;uniqueIndex:idx_equip_code_report_dedupe"`
	ReportDay   uint32 `gorm:"not_null;uniqueIndex:idx_equip_code_report_dedupe"`

	ShipGroupID uint32 `gorm:"not_null"`
	ReportType  uint32 `gorm:"not_null"`
	CreatedAt   time.Time
}
