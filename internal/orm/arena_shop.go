package orm

type ArenaShopState struct {
	CommanderID     uint32 `gorm:"primary_key"`
	FlashCount      uint32 `gorm:"not_null"`
	LastRefreshTime uint32 `gorm:"not_null"`
	NextFlashTime   uint32 `gorm:"not_null"`
}
