package orm

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/ggmolly/belfast/internal/db"
)

type CommanderSkillClass struct {
	CommanderID uint32
	RoomID      uint32
	ShipID      uint32
	SkillPos    uint32
	SkillID     uint32
	StartTime   uint32
	FinishTime  uint32
	Exp         uint32
}

type CommanderShipSkill struct {
	CommanderID uint32
	ShipID      uint32
	SkillPos    uint32
	SkillID     uint32
	Level       uint32
	Exp         uint32
}

var ErrSkillClassConflict = errors.New("skill class conflict")

func ListCommanderSkillClasses(commanderID uint32) ([]CommanderSkillClass, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT commander_id, room_id, ship_id, skill_pos, skill_id, start_time, finish_time, exp
FROM commander_skill_classes
WHERE commander_id = $1
ORDER BY room_id ASC
`, int64(commanderID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]CommanderSkillClass, 0)
	for rows.Next() {
		var entry CommanderSkillClass
		if err := rows.Scan(&entry.CommanderID, &entry.RoomID, &entry.ShipID, &entry.SkillPos, &entry.SkillID, &entry.StartTime, &entry.FinishTime, &entry.Exp); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func GetCommanderSkillClassByRoomTx(ctx context.Context, tx pgx.Tx, commanderID uint32, roomID uint32) (*CommanderSkillClass, error) {
	row := tx.QueryRow(ctx, `
SELECT commander_id, room_id, ship_id, skill_pos, skill_id, start_time, finish_time, exp
FROM commander_skill_classes
WHERE commander_id = $1 AND room_id = $2
FOR UPDATE
`, int64(commanderID), int64(roomID))

	var entry CommanderSkillClass
	if err := row.Scan(&entry.CommanderID, &entry.RoomID, &entry.ShipID, &entry.SkillPos, &entry.SkillID, &entry.StartTime, &entry.FinishTime, &entry.Exp); err != nil {
		return nil, db.MapNotFound(err)
	}
	return &entry, nil
}

func CreateCommanderSkillClassTx(ctx context.Context, tx pgx.Tx, entry *CommanderSkillClass) error {
	_, err := tx.Exec(ctx, `
INSERT INTO commander_skill_classes (commander_id, room_id, ship_id, skill_pos, skill_id, start_time, finish_time, exp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`,
		int64(entry.CommanderID),
		int64(entry.RoomID),
		int64(entry.ShipID),
		int64(entry.SkillPos),
		int64(entry.SkillID),
		int64(entry.StartTime),
		int64(entry.FinishTime),
		int64(entry.Exp),
	)
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrSkillClassConflict
	}
	return err
}

func DeleteCommanderSkillClassTx(ctx context.Context, tx pgx.Tx, commanderID uint32, roomID uint32) error {
	res, err := tx.Exec(ctx, `
DELETE FROM commander_skill_classes
WHERE commander_id = $1 AND room_id = $2
`, int64(commanderID), int64(roomID))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func GetOrCreateCommanderShipSkillTx(ctx context.Context, tx pgx.Tx, commanderID uint32, shipID uint32, skillPos uint32, skillID uint32) (*CommanderShipSkill, error) {
	if _, err := tx.Exec(ctx, `
INSERT INTO commander_ship_skills (commander_id, ship_id, skill_pos, skill_id, level, exp)
VALUES ($1, $2, $3, $4, 1, 0)
ON CONFLICT (commander_id, ship_id, skill_pos)
DO NOTHING
`, int64(commanderID), int64(shipID), int64(skillPos), int64(skillID)); err != nil {
		return nil, err
	}

	row := tx.QueryRow(ctx, `
SELECT commander_id, ship_id, skill_pos, skill_id, level, exp
FROM commander_ship_skills
WHERE commander_id = $1 AND ship_id = $2 AND skill_pos = $3
FOR UPDATE
`, int64(commanderID), int64(shipID), int64(skillPos))

	var entry CommanderShipSkill
	if err := row.Scan(&entry.CommanderID, &entry.ShipID, &entry.SkillPos, &entry.SkillID, &entry.Level, &entry.Exp); err != nil {
		return nil, db.MapNotFound(err)
	}
	if entry.SkillID == 0 {
		entry.SkillID = skillID
	}
	return &entry, nil
}

func SaveCommanderShipSkillTx(ctx context.Context, tx pgx.Tx, entry *CommanderShipSkill) error {
	res, err := tx.Exec(ctx, `
UPDATE commander_ship_skills
SET skill_id = $4, level = $5, exp = $6
WHERE commander_id = $1 AND ship_id = $2 AND skill_pos = $3
`, int64(entry.CommanderID), int64(entry.ShipID), int64(entry.SkillPos), int64(entry.SkillID), int64(entry.Level), int64(entry.Exp))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func GetCommanderShipSkill(commanderID uint32, shipID uint32, skillPos uint32) (*CommanderShipSkill, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT commander_id, ship_id, skill_pos, skill_id, level, exp
FROM commander_ship_skills
WHERE commander_id = $1 AND ship_id = $2 AND skill_pos = $3
`, int64(commanderID), int64(shipID), int64(skillPos))

	var entry CommanderShipSkill
	if err := row.Scan(&entry.CommanderID, &entry.ShipID, &entry.SkillPos, &entry.SkillID, &entry.Level, &entry.Exp); err != nil {
		return nil, db.MapNotFound(err)
	}
	return &entry, nil
}
