package orm

import "time"

type UserAccount struct {
	ID                string     `gorm:"primary_key;size:36"`
	CommanderID       uint32     `gorm:"not_null;uniqueIndex"`
	PasswordHash      string     `gorm:"not_null"`
	PasswordAlgo      string     `gorm:"size:32;not_null"`
	PasswordUpdatedAt time.Time  `gorm:"type:timestamp;not_null"`
	DisabledAt        *time.Time `gorm:"type:timestamp"`
	LastLoginAt       *time.Time `gorm:"type:timestamp"`
	CreatedAt         time.Time  `gorm:"autoCreateTime"`
	UpdatedAt         time.Time  `gorm:"autoUpdateTime"`
}
