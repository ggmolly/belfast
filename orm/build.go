package orm

import (
	"errors"
	"time"

	"github.com/ggmolly/belfast/protobuf"
)

type Build struct {
	ID         uint32    `gorm:"primary_key"`
	BuilderID  uint32    `gorm:"not_null"`
	ShipID     uint32    `gorm:"not_null"`
	FinishesAt time.Time `gorm:"not_null"`

	Ship      Ship      `gorm:"foreignKey:ShipID;references:TemplateID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Commander Commander `gorm:"foreignKey:BuilderID;references:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

var (
	ErrorNotEnoughQuickFinishers = errors.New("not enough quick finishers")
)

// Inserts or updates a build in the database (based on the primary key)
func (b *Build) Create() error {
	return GormDB.Save(b).Error
}

// Updates a build in the database
func (b *Build) Update() error {
	return GormDB.Model(b).Updates(b).Error
}

// Gets a build from the database by its primary key
// If greedy is true, it will also load the relations
func (b *Build) Retrieve(greedy bool) error {
	if greedy {
		return GormDB.
			Joins("JOIN ships ON ships.template_id = builds.ship_id").
			Joins("JOIN commanders ON commanders.commander_id = builds.builder_id").
			Where("builds.id = ?", b.ID).
			First(b).Error
	} else {
		return GormDB.
			Where("id = ?", b.ID).
			First(b).Error
	}
}

// Deletes a build from the database
func (b *Build) Delete() error {
	return GormDB.Delete(b).Error
}

// Removes the build from the database and adds the ship to the commander
func (b *Build) Consume(shipId uint32, commander *Commander) (*protobuf.SHIPINFO, error) {
	// Delete the build in the database
	if err := b.Delete(); err != nil {
		return nil, err
	}
	// Delete the build in the commander
	for i, build := range commander.Builds {
		if build.ID == b.ID {
			commander.Builds = append(commander.Builds[:i], commander.Builds[i+1:]...)
			break
		}
	}
	// Add the ship to the commander
	ship, err := commander.AddShip(shipId)
	if err != nil {
		return nil, err
	}
	commander.IncrementExchangeCount(uint32(len(commander.Builds)))
	return &ship, nil
}

// QuickFinishes a build, checks if the passed commander has enough quick finishers
func (b *Build) QuickFinish(commander *Commander) error {
	if !commander.HasEnoughItem(15003, 1) {
		return ErrorNotEnoughQuickFinishers
	}
	b.FinishesAt = time.Now().Add(-1 * time.Second)
	if err := b.Update(); err != nil {
		return err
	}
	return commander.ConsumeItem(15003, 1)
}
