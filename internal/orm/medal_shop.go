package orm

type MedalShopState struct {
	CommanderID     uint32 `gorm:"primary_key"`
	NextRefreshTime uint32 `gorm:"not_null"`
}

type MedalShopGood struct {
	CommanderID uint32 `gorm:"primary_key"`
	Index       uint32 `gorm:"primary_key"`
	GoodsID     uint32 `gorm:"not_null"`
	Count       uint32 `gorm:"not_null"`
}
