package orm

import "time"

type UserRegistrationChallenge struct {
	ID           string     `gorm:"primary_key;size:36"`
	CommanderID  uint32     `gorm:"not_null;index"`
	Pin          string     `gorm:"size:6;not_null;index"`
	PasswordHash string     `gorm:"not_null"`
	PasswordAlgo string     `gorm:"size:32;not_null"`
	Status       string     `gorm:"size:16;not_null;index"`
	ExpiresAt    time.Time  `gorm:"type:timestamp;not_null"`
	ConsumedAt   *time.Time `gorm:"type:timestamp"`
	CreatedAt    time.Time  `gorm:"autoCreateTime"`
}
