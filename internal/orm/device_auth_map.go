package orm

import "time"

type DeviceAuthMap struct {
	DeviceID  string    `gorm:"primary_key;size:128"`
	Arg2      uint32    `gorm:"not_null"`
	AccountID uint32    `gorm:"not_null"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
