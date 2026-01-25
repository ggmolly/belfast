package orm

import "time"

type RemasterProgress struct {
	ID          uint64    `gorm:"primaryKey"`
	CommanderID uint32    `gorm:"not_null;index;uniqueIndex:idx_remaster_progress"`
	ChapterID   uint32    `gorm:"not_null;index;uniqueIndex:idx_remaster_progress"`
	Pos         uint32    `gorm:"not_null;uniqueIndex:idx_remaster_progress"`
	Count       uint32    `gorm:"not_null;default:0"`
	Received    bool      `gorm:"not_null;default:false"`
	CreatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
	UpdatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
}
