package orm

type CommanderItem struct {
	CommanderID uint32 `gorm:"not_null;primaryKey"`
	ItemID      uint32 `gorm:"not_null;primaryKey"`
	Count       uint32 `gorm:"not_null"`

	Commander Commander `gorm:"foreignKey:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Item      Item      `gorm:"foreignKey:ItemID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

// type CommanderLimitItem struct {
// 	CommanderID uint32 `gorm:"not_null;primaryKey"`
// 	ItemID      uint32 `gorm:"not_null;primaryKey"`
// 	Count       uint32 `gorm:"not_null"`

// 	Commander Commander `gorm:"foreignKey:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
// 	Item      Item      `gorm:"foreignKey:ItemID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
// }

type CommanderMiscItem struct {
	CommanderID uint32 `gorm:"not_null;primaryKey"`
	ItemID      uint32 `gorm:"not_null;primaryKey"`
	Data        uint32 `gorm:"not_null;primaryKey"`

	Commander Commander `gorm:"foreignKey:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Item      Item      `gorm:"foreignKey:ItemID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
