package orm

import "github.com/lib/pq"

type ShopOffer struct {
	ID             uint32        `gorm:"primary_key"`
	Effects        pq.Int64Array `gorm:"type:integer[];not_null"`
	Number         uint32        `gorm:"not_null"`
	ResourceNumber uint32        `gorm:"not_null"`
	ResourceID     uint32        `gorm:"not_null"`
	Type           uint32        `gorm:"not_null"`

	Resource Resource `gorm:"foreignkey:ResourceID;references:ID"`
}
