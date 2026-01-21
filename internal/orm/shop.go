package orm

import "encoding/json"

type ShopOffer struct {
	ID             uint32          `gorm:"primary_key" json:"id"`
	Effects        Int64List       `gorm:"type:json;not_null" json:"-"`
	EffectArgs     json.RawMessage `gorm:"type:json" json:"effect_args"`
	Number         int             `gorm:"not_null" json:"num"`
	ResourceNumber int             `gorm:"not_null" json:"resource_num"`
	ResourceID     uint32          `gorm:"not_null" json:"resource_type"`
	Type           uint32          `gorm:"not_null" json:"type"`

	Resource Resource `gorm:"foreignkey:ResourceID;references:ID"`
}
