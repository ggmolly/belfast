package orm

import (
	"context"
	"time"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

type CommanderBuff struct {
	CommanderID uint32    `gorm:"primaryKey;autoIncrement:false"`
	BuffID      uint32    `gorm:"primaryKey;autoIncrement:false"`
	ExpiresAt   time.Time `gorm:"not_null;index:idx_commander_buff_expires_at"`
}

func ListCommanderBuffs(commanderID uint32) ([]CommanderBuff, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListCommanderBuffs(ctx, int64(commanderID))
	if err != nil {
		return nil, err
	}
	buffs := make([]CommanderBuff, 0, len(rows))
	for _, r := range rows {
		buffs = append(buffs, CommanderBuff{CommanderID: uint32(r.CommanderID), BuffID: uint32(r.BuffID), ExpiresAt: r.ExpiresAt.Time})
	}
	return buffs, nil
}

func UpsertCommanderBuff(commanderID uint32, buffID uint32, expiresAt time.Time) error {
	ctx := context.Background()
	return db.DefaultStore.Queries.UpsertCommanderBuff(ctx, gen.UpsertCommanderBuffParams{CommanderID: int64(commanderID), BuffID: int64(buffID), ExpiresAt: pgTimestamptz(expiresAt.UTC())})
}

func ListCommanderActiveBuffs(commanderID uint32, now time.Time) ([]CommanderBuff, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListCommanderActiveBuffs(ctx, gen.ListCommanderActiveBuffsParams{CommanderID: int64(commanderID), ExpiresAt: pgTimestamptz(now.UTC())})
	if err != nil {
		return nil, err
	}
	buffs := make([]CommanderBuff, 0, len(rows))
	for _, r := range rows {
		buffs = append(buffs, CommanderBuff{CommanderID: uint32(r.CommanderID), BuffID: uint32(r.BuffID), ExpiresAt: r.ExpiresAt.Time})
	}
	return buffs, nil
}

func GetCommanderBuff(commanderID uint32, buffID uint32) (*CommanderBuff, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT commander_id, buff_id, expires_at
FROM commander_buffs
WHERE commander_id = $1
  AND buff_id = $2
`, int64(commanderID), int64(buffID))

	buff := CommanderBuff{}
	err := row.Scan(&buff.CommanderID, &buff.BuffID, &buff.ExpiresAt)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &buff, nil
}

func UpdateCommanderBuffExpiry(commanderID uint32, buffID uint32, expiresAt time.Time) error {
	ctx := context.Background()
	res, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE commander_buffs
SET expires_at = $3
WHERE commander_id = $1
  AND buff_id = $2
`, int64(commanderID), int64(buffID), expiresAt.UTC())
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func DeleteCommanderBuff(commanderID uint32, buffID uint32) error {
	ctx := context.Background()
	res, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM commander_buffs
WHERE commander_id = $1
  AND buff_id = $2
`, int64(commanderID), int64(buffID))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}
