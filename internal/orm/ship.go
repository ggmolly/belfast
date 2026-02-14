package orm

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
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
	ctx := context.Background()
	return db.DefaultStore.Queries.UpsertShipRecord(ctx, gen.UpsertShipRecordParams{
		TemplateID:  int64(s.TemplateID),
		Name:        s.Name,
		EnglishName: s.EnglishName,
		RarityID:    int64(s.RarityID),
		Star:        int64(s.Star),
		Type:        int64(s.Type),
		Nationality: int64(s.Nationality),
		BuildTime:   int64(s.BuildTime),
		PoolID:      pgInt8FromUint32Ptr(s.PoolID),
	})
}

func InsertShip(s *Ship) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO ships (
	template_id,
	name,
	english_name,
	rarity_id,
	star,
	type,
	nationality,
	build_time,
	pool_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`,
		int64(s.TemplateID),
		s.Name,
		s.EnglishName,
		int64(s.RarityID),
		int64(s.Star),
		int64(s.Type),
		int64(s.Nationality),
		int64(s.BuildTime),
		pgInt8FromUint32Ptr(s.PoolID),
	)
	return err
}

// Updates a ship in the database
func (s *Ship) Update() error {
	return s.Create()
}

// Gets a ship from the database by its primary key
// If greedy is true, it will also load the relations
func (s *Ship) Retrieve(greedy bool) error {
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetShip(ctx, int64(s.TemplateID))
	err = db.MapNotFound(err)
	if err != nil {
		return err
	}
	_ = greedy
	s.Name = row.Name
	s.EnglishName = row.EnglishName
	s.RarityID = uint32(row.RarityID)
	s.Star = uint32(row.Star)
	s.Type = uint32(row.Type)
	s.Nationality = uint32(row.Nationality)
	s.BuildTime = uint32(row.BuildTime)
	s.PoolID = pgInt8PtrToUint32Ptr(row.PoolID)
	return nil
}

// Deletes a ship from the database
func (s *Ship) Delete() error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Queries.DeleteShip(ctx, int64(s.TemplateID))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func ValidateShipID(shipID uint32) error {
	ctx := context.Background()
	count, err := db.DefaultStore.Queries.CountShipByTemplateID(ctx, int64(shipID))
	if err != nil {
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
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetRandomPoolShip(ctx, gen.GetRandomPoolShipParams{PoolID: pgtype.Int8{Int64: int64(poolId), Valid: true}, RarityID: int64(rarity)})
	err = db.MapNotFound(err)
	if err != nil {
		return Ship{}, err
	}
	randomShip := Ship{
		TemplateID:  uint32(row.TemplateID),
		Name:        row.Name,
		EnglishName: row.EnglishName,
		RarityID:    uint32(row.RarityID),
		Star:        uint32(row.Star),
		Type:        uint32(row.Type),
		Nationality: uint32(row.Nationality),
		BuildTime:   uint32(row.BuildTime),
		PoolID:      pgInt8PtrToUint32Ptr(row.PoolID),
	}
	return randomShip, nil
}
