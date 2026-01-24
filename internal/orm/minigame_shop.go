package orm

type MiniGameShopState struct {
	CommanderID     uint32 `gorm:"primary_key"`
	NextRefreshTime uint32 `gorm:"not_null"`
}

type MiniGameShopGood struct {
	CommanderID uint32 `gorm:"primary_key"`
	GoodsID     uint32 `gorm:"primary_key"`
	Count       uint32 `gorm:"not_null"`
}
