package orm

type ShoppingStreetState struct {
	CommanderID   uint32 `gorm:"primary_key"`
	Level         uint32 `gorm:"not_null"`
	NextFlashTime uint32 `gorm:"not_null"`
	LevelUpTime   uint32 `gorm:"not_null"`
	FlashCount    uint32 `gorm:"not_null"`
}

type ShoppingStreetGood struct {
	CommanderID uint32 `gorm:"primary_key"`
	GoodsID     uint32 `gorm:"primary_key"`
	Discount    uint32 `gorm:"not_null"`
	BuyCount    uint32 `gorm:"not_null"`
}
