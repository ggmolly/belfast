package orm

type ShopOffer struct {
	ID             uint32    `gorm:"primary_key" json:"id"`
	Effects        Int64List `gorm:"type:json;not_null" json:"-"`
	Number         uint32    `gorm:"not_null" json:"num"`
	ResourceNumber uint32    `gorm:"not_null" json:"resource_num"`
	ResourceID     uint32    `gorm:"not_null" json:"resource_type"`
	Type           uint32    `gorm:"not_null" json:"type"`

	Resource Resource `gorm:"foreignkey:ResourceID;references:ID"`
}
