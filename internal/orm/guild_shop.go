package orm

type GuildShopState struct {
	CommanderID     uint32 `gorm:"primary_key"`
	RefreshCount    uint32 `gorm:"not_null"`
	NextRefreshTime uint32 `gorm:"not_null"`
}

type GuildShopGood struct {
	CommanderID uint32 `gorm:"primary_key"`
	Index       uint32 `gorm:"primary_key"`
	GoodsID     uint32 `gorm:"not_null"`
	Count       uint32 `gorm:"not_null"`
}
