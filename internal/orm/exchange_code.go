package orm

import (
	"encoding/json"
	"time"
)

type ExchangeCode struct {
	ID        uint32          `gorm:"primary_key"`
	Code      string          `gorm:"size:64;not_null;uniqueIndex"`
	Platform  string          `gorm:"size:32"`
	Quota     int             `gorm:"not_null;default:-1"`
	Rewards   json.RawMessage `gorm:"type:json;not_null"`
	CreatedAt time.Time       `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
	UpdatedAt time.Time       `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
}

type ExchangeCodeRedeem struct {
	ExchangeCodeID uint32    `gorm:"primaryKey"`
	CommanderID    uint32    `gorm:"primaryKey"`
	RedeemedAt     time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
}
