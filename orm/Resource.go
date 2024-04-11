package orm

var (
	resourceAliases = map[uint32]uint32{
		14: 4, // freeGem <=> gem
	}
)

type OwnedResource struct {
	CommanderID uint32 `gorm:"primaryKey"`
	ResourceID  uint32 `gorm:"primaryKey"`
	Amount      uint32 `gorm:"not null;default:0"`

	Commander Commander `gorm:"foreignKey:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Resource  Resource  `gorm:"foreignKey:ResourceID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type Resource struct {
	ID     uint32 `gorm:"primary_key"`
	ItemID uint32
	Name   string `gorm:"type:varchar(50);not null"`

	Item Item `gorm:"foreignKey:ItemID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

// Some resources are aliases, for example id=14 = freeGem <=> id=4 = gem
func DealiasResource(resourceId *uint32) {
	if alias, ok := resourceAliases[*resourceId]; ok {
		*resourceId = alias
	}
}
