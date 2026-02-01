package orm

import "time"

type LocalAccount struct {
	Arg2      uint32    `gorm:"primary_key;not_null;uniqueIndex"`
	Account   string    `gorm:"size:128;not_null;uniqueIndex"`
	Password  string    `gorm:"size:256;not_null"`
	MailBox   string    `gorm:"size:256;not_null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
