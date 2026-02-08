package orm

import "time"

// Account is the single authenticated principal for the REST API.
//
// It can represent:
// - A staff/admin account (Username set, IsAdmin true/false)
// - A player-bound account (CommanderID set)
// - Both (rare, but supported)
type Account struct {
	ID string `gorm:"primary_key;size:36"`

	Username           *string `gorm:"size:64"`
	UsernameNormalized *string `gorm:"size:64;uniqueIndex"`
	CommanderID        *uint32 `gorm:"uniqueIndex"`

	PasswordHash      string    `gorm:"not_null"`
	PasswordAlgo      string    `gorm:"size:32;not_null"`
	PasswordUpdatedAt time.Time `gorm:"type:timestamp;not_null"`

	IsAdmin     bool       `gorm:"default:false;not_null"`
	DisabledAt  *time.Time `gorm:"type:timestamp"`
	LastLoginAt *time.Time `gorm:"type:timestamp"`

	WebAuthnUserHandle []byte `gorm:"uniqueIndex;type:blob"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
