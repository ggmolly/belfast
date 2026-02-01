package orm

import "time"

type AdminAuditLog struct {
	ID           string    `gorm:"primary_key;size:36"`
	ActorUserID  *string   `gorm:"size:36;index"`
	Action       string    `gorm:"size:64;not_null;index"`
	TargetUserID *string   `gorm:"size:36;index"`
	Metadata     []byte    `gorm:"type:json"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}
