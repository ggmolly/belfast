package orm

import "time"

type Debug struct {
	FrameID    uint      `gorm:"primary_key"`
	PacketSize int       `gorm:"not_null"`
	PacketID   int       `gorm:"not_null"`
	Data       []byte    `gorm:"not_null"`
	LoggedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP"`

	DebugName DebugName `gorm:"foreignKey:PacketID"`
}

type DebugName struct {
	ID   int    `gorm:"primary_key"`
	Name string `gorm:"default:'Unknown'"`

	Debug []Debug `gorm:"foreignKey:PacketID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
