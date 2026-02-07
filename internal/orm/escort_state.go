package orm

import (
	"encoding/json"
	"time"
)

type EscortState struct {
	ID             uint64          `gorm:"primary_key"`
	AccountID      uint32          `gorm:"not_null;index:idx_escort_account_line,unique"`
	LineID         uint32          `gorm:"not_null;index:idx_escort_account_line,unique"`
	AwardTimestamp uint32          `gorm:"not_null;default:0"`
	FlashTimestamp uint32          `gorm:"not_null;default:0"`
	MapPositions   json.RawMessage `gorm:"type:json"`
	CreatedAt      time.Time       `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
	UpdatedAt      time.Time       `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
}
