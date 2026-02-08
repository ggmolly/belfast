package orm

import "time"

type AuditLog struct {
	ID string `gorm:"primary_key;size:36"`

	ActorAccountID   *string `gorm:"size:36;index"`
	ActorCommanderID *uint32 `gorm:"index"`

	Method     string `gorm:"size:8;not_null;index"`
	Path       string `gorm:"size:255;not_null;index"`
	StatusCode int    `gorm:"not_null;index"`

	PermissionKey *string `gorm:"size:128;index"`
	PermissionOp  *string `gorm:"size:16;index"`
	Action        string  `gorm:"size:96;index"`

	Metadata  []byte    `gorm:"type:json"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
