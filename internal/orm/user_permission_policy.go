package orm

import "time"

type UserPermissionPolicy struct {
	ID        string    `gorm:"primary_key;size:36"`
	Actions   []byte    `gorm:"type:json"`
	UpdatedBy *string   `gorm:"size:36"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
