package orm

import (
	"context"
	"time"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

type Punishment struct {
	ID            uint32     `gorm:"primary_key"`
	PunishedID    uint32     `gorm:"not_null"`
	LiftTimestamp *time.Time `gorm:"type:timestamp"`
	IsPermanent   bool       `gorm:"default:false;not_null"`

	Punished Commander `gorm:"foreignKey:PunishedID;references:CommanderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

// Inserts or updates a punishment in the database (based on the primary key)
func (p *Punishment) Create() error {
	ctx := context.Background()
	id, err := db.DefaultStore.Queries.UpsertPunishment(ctx, gen.UpsertPunishmentParams{
		PunishedID:    int64(p.PunishedID),
		LiftTimestamp: pgTimestamptzFromPtr(p.LiftTimestamp),
		IsPermanent:   p.IsPermanent,
	})
	if err != nil {
		return err
	}
	p.ID = uint32(id)
	return nil
}

// Updates a punishment in the database
func (p *Punishment) Update() error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Queries.UpdatePunishment(ctx, gen.UpdatePunishmentParams{
		ID:            int64(p.ID),
		PunishedID:    int64(p.PunishedID),
		LiftTimestamp: pgTimestamptzFromPtr(p.LiftTimestamp),
		IsPermanent:   p.IsPermanent,
	})
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

// Gets a punishment from the database by its primary key
// If greedy is true, it will also load the relations
func (p *Punishment) Retrieve(greedy bool) error {
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetPunishment(ctx, int64(p.ID))
	err = db.MapNotFound(err)
	if err != nil {
		return err
	}
	p.PunishedID = uint32(row.PunishedID)
	p.LiftTimestamp = pgTimestamptzPtr(row.LiftTimestamp)
	p.IsPermanent = row.IsPermanent
	_ = greedy
	return nil
}

// Deletes a punishment from the database
func (p *Punishment) Delete() error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Queries.DeletePunishment(ctx, int64(p.ID))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func ListPunishmentsByCommanderID(commanderID uint32) ([]Punishment, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT id, punished_id, lift_timestamp, is_permanent
FROM punishments
WHERE punished_id = $1
ORDER BY id DESC
`, int64(commanderID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	punishments := make([]Punishment, 0)
	for rows.Next() {
		punishment, err := scanPunishment(rows)
		if err != nil {
			return nil, err
		}
		punishments = append(punishments, *punishment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return punishments, nil
}

func GetPunishmentByCommanderAndID(commanderID uint32, punishmentID uint32) (*Punishment, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT id, punished_id, lift_timestamp, is_permanent
FROM punishments
WHERE punished_id = $1 AND id = $2
`, int64(commanderID), int64(punishmentID))
	punishment, err := scanPunishment(row)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return punishment, nil
}

func DeletePunishmentByCommanderAndID(commanderID uint32, punishmentID uint32) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM punishments
WHERE punished_id = $1 AND id = $2
`, int64(commanderID), int64(punishmentID))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func ListPermanentPunishmentsByCommanderIDs(commanderIDs []uint32) ([]Punishment, error) {
	if len(commanderIDs) == 0 {
		return []Punishment{}, nil
	}
	ids := make([]int64, 0, len(commanderIDs))
	for _, id := range commanderIDs {
		ids = append(ids, int64(id))
	}

	ctx := context.Background()
	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT id, punished_id, lift_timestamp, is_permanent
FROM punishments
WHERE punished_id = ANY($1::bigint[])
  AND lift_timestamp IS NULL
`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	punishments := make([]Punishment, 0)
	for rows.Next() {
		punishment, scanErr := scanPunishment(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		punishments = append(punishments, *punishment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return punishments, nil
}

type punishmentScanner interface {
	Scan(dest ...any) error
}

func scanPunishment(scanner punishmentScanner) (*Punishment, error) {
	var punishment Punishment
	var liftTimestamp *time.Time
	if err := scanner.Scan(&punishment.ID, &punishment.PunishedID, &liftTimestamp, &punishment.IsPermanent); err != nil {
		return nil, err
	}
	punishment.LiftTimestamp = liftTimestamp
	return &punishment, nil
}
