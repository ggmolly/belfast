package orm

import "time"

type WebAuthnCredential struct {
	ID             string     `gorm:"primary_key;size:36"`
	UserID         string     `gorm:"size:36;not_null;index"`
	CredentialID   string     `gorm:"not_null;uniqueIndex"`
	PublicKey      []byte     `gorm:"not_null"`
	SignCount      uint32     `gorm:"not_null"`
	Transports     StringList `gorm:"type:json"`
	AAGUID         string     `gorm:"size:64"`
	AttestationFmt string     `gorm:"size:64"`
	ResidentKey    bool       `gorm:"default:false;not_null"`
	BackupEligible *bool      `gorm:""`
	BackupState    *bool      `gorm:""`
	CreatedAt      time.Time  `gorm:"autoCreateTime"`
	LastUsedAt     *time.Time `gorm:"type:timestamp"`
	Label          *string    `gorm:"size:128"`
	RPID           string     `gorm:"size:255;not_null"`
}
