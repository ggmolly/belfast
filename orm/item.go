package orm

type Item struct {
	ID          uint32 `gorm:"primary_key" json:"id"`
	Name        string `gorm:"size:70;not_null" json:"name"`
	Rarity      int    `gorm:"not_null" json:"rarity"`
	ShopID      int    `gorm:"not_null;default:-2" json:"shop_id,omitempty"`
	Type        int    `gorm:"not_null" json:"type"`
	VirtualType int    `gorm:"not_null" json:"virtual_type"`
}
