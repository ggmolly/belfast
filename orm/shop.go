package orm

import "github.com/lib/pq"

type ShopOffer struct {
	ID             uint32        `gorm:"primary_key" json:"id"`
	Effects        pq.Int64Array `gorm:"type:integer[];not_null" json:"-"`
	Number         uint32        `gorm:"not_null" json:"num"`
	ResourceNumber uint32        `gorm:"not_null" json:"resource_num"`
	ResourceID     uint32        `gorm:"not_null" json:"resource_type"`
	Type           uint32        `gorm:"not_null" json:"type"`

	Resource Resource `gorm:"foreignkey:ResourceID;references:ID"`
}
