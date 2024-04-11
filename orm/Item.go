package orm

type Item struct {
	ID          uint32 `gorm:"primary_key"`
	Name        string `gorm:"size:70;not null"`
	Rarity      int    `gorm:"not null"`
	ShopID      int    `gorm:"not null"`
	Type        int    `gorm:"not null"`
	VirtualType int    `gorm:"not null"`
}
