package orm

import "time"

type RemasterState struct {
	CommanderID      uint32    `gorm:"primaryKey;autoIncrement:false"`
	TicketCount      uint32    `gorm:"not_null;default:0"`
	ActiveChapterID  uint32    `gorm:"not_null;default:0"`
	DailyCount       uint32    `gorm:"not_null;default:0"`
	LastDailyResetAt time.Time `gorm:"type:timestamp;default:'1970-01-01 00:00:00';not_null"`
	CreatedAt        time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
	UpdatedAt        time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
}
