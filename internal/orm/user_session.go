package orm

import "time"

type UserSession struct {
	ID            string     `gorm:"primary_key;size:36"`
	UserID        string     `gorm:"size:36;not_null;index"`
	CreatedAt     time.Time  `gorm:"autoCreateTime"`
	LastSeenAt    time.Time  `gorm:"type:timestamp;not_null"`
	ExpiresAt     time.Time  `gorm:"type:timestamp;not_null"`
	IPAddress     string     `gorm:"size:64"`
	UserAgent     string     `gorm:"size:255"`
	RevokedAt     *time.Time `gorm:"type:timestamp"`
	CSRFToken     string     `gorm:"size:64;not_null"`
	CSRFExpiresAt time.Time  `gorm:"type:timestamp;not_null"`
}
