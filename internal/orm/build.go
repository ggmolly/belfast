package orm

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ggmolly/belfast/internal/db"
)

type Build struct {
	ID         uint32    `gorm:"primary_key"`
	BuilderID  uint32    `gorm:"not_null"`
	ShipID     uint32    `gorm:"not_null"`
	PoolID     uint32    `gorm:"not_null"`
	FinishesAt time.Time `gorm:"not_null"`

	Ship      Ship      `gorm:"foreignKey:ShipID;references:TemplateID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Commander Commander `gorm:"foreignKey:BuilderID;references:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

var (
	ErrorNotEnoughQuickFinishers = errors.New("not enough quick finishers")
)

// Inserts or updates a build in the database (based on the primary key)
func (b *Build) Create() error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
INSERT INTO builds (builder_id, ship_id, pool_id, finishes_at)
VALUES ($1, $2, $3, $4)
RETURNING id
`, int64(b.BuilderID), int64(b.ShipID), int64(b.PoolID), b.FinishesAt)
	var id int64
	if err := row.Scan(&id); err != nil {
		return err
	}
	b.ID = uint32(id)
	return nil
}

// Updates a build in the database
func (b *Build) Update() error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE builds
SET builder_id = $2, ship_id = $3, pool_id = $4, finishes_at = $5
WHERE id = $1
`, int64(b.ID), int64(b.BuilderID), int64(b.ShipID), int64(b.PoolID), b.FinishesAt)
	return err
}

// Gets a build from the database by its primary key
// If greedy is true, it will also load the relations
func (b *Build) Retrieve(greedy bool) error {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `SELECT id, builder_id, ship_id, pool_id, finishes_at FROM builds WHERE id = $1`, int64(b.ID))
	var id int64
	var builderID int64
	var shipID int64
	var poolID int64
	if err := row.Scan(&id, &builderID, &shipID, &poolID, &b.FinishesAt); err != nil {
		return db.MapNotFound(err)
	}
	b.ID = uint32(id)
	b.BuilderID = uint32(builderID)
	b.ShipID = uint32(shipID)
	b.PoolID = uint32(poolID)
	if greedy {
		ship := Ship{TemplateID: b.ShipID}
		if err := ship.Retrieve(false); err != nil {
			return err
		}
		b.Ship = ship

		commander, err := GetCommanderCoreByID(b.BuilderID)
		if err != nil {
			return err
		}
		b.Commander = *commander
	}
	return nil
}

// Deletes a build from the database
func (b *Build) Delete() error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	res, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM builds WHERE id = $1`, int64(b.ID))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func GetBuildByID(buildID uint32) (*Build, error) {
	if db.DefaultStore == nil {
		return nil, errors.New("db not initialized")
	}
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `SELECT id, builder_id, ship_id, pool_id, finishes_at FROM builds WHERE id = $1`, int64(buildID))
	var build Build
	var id int64
	var builderID int64
	var shipID int64
	var poolID int64
	if err := row.Scan(&id, &builderID, &shipID, &poolID, &build.FinishesAt); err != nil {
		err = db.MapNotFound(err)
		return nil, err
	}
	build.ID = uint32(id)
	build.BuilderID = uint32(builderID)
	build.ShipID = uint32(shipID)
	build.PoolID = uint32(poolID)
	return &build, nil
}

// Removes the build from the database and adds the ship to the commander
func (b *Build) Consume(shipId uint32, commander *Commander) (*OwnedShip, error) {
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
	return ship, nil
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

var _ = pgx.ErrNoRows
