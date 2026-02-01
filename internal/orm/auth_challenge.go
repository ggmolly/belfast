package orm

import "time"

type AuthChallenge struct {
	ID        string    `gorm:"primary_key;size:36"`
	UserID    *string   `gorm:"size:36;index"`
	Type      string    `gorm:"size:64;not_null;index"`
	Challenge string    `gorm:"size:512;not_null;index"`
	ExpiresAt time.Time `gorm:"type:timestamp;not_null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	Metadata  []byte    `gorm:"type:json"`
}
