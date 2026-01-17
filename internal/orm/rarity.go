package orm

type Rarity struct {
	ID   uint32 `gorm:"primary_key"`
	Name string `gorm:"size:12"`
}
