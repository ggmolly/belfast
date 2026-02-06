package orm

import "time"

type Role struct {
	ID          string    `gorm:"primary_key;size:36"`
	Name        string    `gorm:"size:64;not_null;uniqueIndex"`
	Description string    `gorm:"size:255;not_null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
	UpdatedBy   *string   `gorm:"size:36"`
}

type Permission struct {
	ID          string    `gorm:"primary_key;size:36"`
	Key         string    `gorm:"size:128;not_null;uniqueIndex"`
	Description string    `gorm:"size:255;not_null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

type RolePermission struct {
	RoleID       string `gorm:"primaryKey;size:36"`
	PermissionID string `gorm:"primaryKey;size:36"`

	CanReadSelf  bool `gorm:"default:false;not_null"`
	CanReadAny   bool `gorm:"default:false;not_null"`
	CanWriteSelf bool `gorm:"default:false;not_null"`
	CanWriteAny  bool `gorm:"default:false;not_null"`

	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type AccountRole struct {
	AccountID string    `gorm:"primaryKey;size:36"`
	RoleID    string    `gorm:"primaryKey;size:36"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

const (
	PermissionOverrideAllow = "allow"
	PermissionOverrideDeny  = "deny"
)

type AccountPermissionOverride struct {
	AccountID    string `gorm:"primaryKey;size:36"`
	PermissionID string `gorm:"primaryKey;size:36"`

	Mode string `gorm:"size:8;not_null"`

	CanReadSelf  bool `gorm:"default:false;not_null"`
	CanReadAny   bool `gorm:"default:false;not_null"`
	CanWriteSelf bool `gorm:"default:false;not_null"`
	CanWriteAny  bool `gorm:"default:false;not_null"`

	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
