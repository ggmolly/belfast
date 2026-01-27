package orm

import (
	"errors"

	"github.com/ggmolly/belfast/internal/rng"
)

type Ship struct {
	TemplateID  uint32 `gorm:"primary_key" json:"id"`
	Name        string `gorm:"size:32;not_null" json:"name"`
	EnglishName string `gorm:"size:64;not_null" json:"english_name"`
	RarityID    uint32 `gorm:"not_null" json:"rarity"`
	Star        uint32 `gorm:"not_null" json:"star"`
	Type        uint32 `gorm:"not_null" json:"type"`
	Nationality uint32 `gorm:"not_null" json:"nationality"`
	BuildTime   uint32 `gorm:"not_null" json:"-"`
	PoolID      *uint32

	// Rarity   Rarity   `gorm:"foreignKey:RarityID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	// ShipType ShipType `gorm:"foreignKey:Type;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type ShipType struct {
	ID   uint32 `gorm:"primary_key"`
	Name string `gorm:"size:32;not_null"`
}

// Inserts or updates a ship in the database (based on the primary key)
func (s *Ship) Create() error {
	return GormDB.Save(s).Error
}

// Updates a ship in the database
func (s *Ship) Update() error {
	return GormDB.Model(s).Updates(s).Error
}

// Gets a ship from the database by its primary key
// If greedy is true, it will also load the relations
func (s *Ship) Retrieve(greedy bool) error {
	if greedy {
		return GormDB.
			Joins("JOIN rarities ON rarities.id = ships.rarity_id").
			Where("ships.template_id = ?", s.TemplateID).
			First(s).Error
	}
	return GormDB.
		Where("template_id = ?", s.TemplateID).
		First(s).Error
}

// Deletes a ship from the database
func (s *Ship) Delete() error {
	return GormDB.Delete(s).Error
}

func ValidateShipID(shipID uint32) error {
	var count int64
	if err := GormDB.Model(&Ship{}).Where("template_id = ?", shipID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("ship not found")
	}
	return nil
}

var (
	shipRng = rng.NewLockedRand()
)

// Returns a random ship from a pool, based on Azur Lane's rates
// 7% Super Rare (gold color)
// 12% Elite (purple color)
// 51% Rare (blue color)
// 30% Common (gray color)
// Azur Lane has sometime some boosted rates for some ships, but we don't care about that for now
func GetRandomPoolShip(poolId uint32) (Ship, error) {
	randomN := shipRng.Uint32N(100) + 1 // between 1 and 100
	var rarity uint32
	if randomN <= 7 {
		rarity = 5 // SR
	} else if randomN <= 19 {
		rarity = 4 // Elite
	} else if randomN <= 70 {
		rarity = 3 // Rare
	} else {
		rarity = 2 // Common
	}
	var randomShip Ship
	err := GormDB.
		Where("pool_id = ? AND rarity_id = ?", poolId, rarity).
		Order("RANDOM()").
		First(&randomShip).Error
	return randomShip, err
}
