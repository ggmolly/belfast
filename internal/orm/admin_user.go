package orm

import "time"

type AdminUser struct {
	ID                 string     `gorm:"primary_key;size:36"`
	Username           string     `gorm:"size:64;not_null;uniqueIndex"`
	UsernameNormalized string     `gorm:"size:64;not_null;uniqueIndex"`
	PasswordHash       string     `gorm:"not_null"`
	PasswordAlgo       string     `gorm:"size:32;not_null"`
	PasswordUpdatedAt  time.Time  `gorm:"type:timestamp;not_null"`
	IsAdmin            bool       `gorm:"default:true;not_null"`
	DisabledAt         *time.Time `gorm:"type:timestamp"`
	LastLoginAt        *time.Time `gorm:"type:timestamp"`
	WebAuthnUserHandle []byte     `gorm:"uniqueIndex;type:blob"`
	CreatedAt          time.Time  `gorm:"autoCreateTime"`
	UpdatedAt          time.Time  `gorm:"autoUpdateTime"`
}
