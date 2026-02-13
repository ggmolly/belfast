package orm

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/ggmolly/belfast/internal/db"
)

// CommanderStoreupAwardProgress tracks which collection (storeup) award tiers a commander has claimed.
// LastAwardIndex is a 1-based index into storeup_data_template.award_display/level.
type CommanderStoreupAwardProgress struct {
	CommanderID    uint32 `gorm:"primaryKey;autoIncrement:false"`
	StoreupID      uint32 `gorm:"primaryKey;autoIncrement:false"`
	LastAwardIndex uint32 `gorm:"not null;default:0"`
}

func GetCommanderStoreupAwardProgress(commanderID uint32, storeupID uint32) (*CommanderStoreupAwardProgress, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT commander_id, storeup_id, last_award_index
FROM commander_storeup_award_progresses
WHERE commander_id = $1 AND storeup_id = $2
`, int64(commanderID), int64(storeupID))
	var out CommanderStoreupAwardProgress
	err := row.Scan(&out.CommanderID, &out.StoreupID, &out.LastAwardIndex)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func GetLastCommanderStoreupAwardIndex(commanderID uint32, storeupID uint32) (uint32, error) {
	row, err := GetCommanderStoreupAwardProgress(commanderID, storeupID)
	if err == nil {
		return row.LastAwardIndex, nil
	}
	if db.IsNotFound(err) {
		return 0, nil
	}
	return 0, err
}

func ListCommanderStoreupAwardProgress(commanderID uint32) ([]CommanderStoreupAwardProgress, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT commander_id, storeup_id, last_award_index
FROM commander_storeup_award_progresses
WHERE commander_id = $1
ORDER BY storeup_id ASC
`, int64(commanderID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]CommanderStoreupAwardProgress, 0)
	for rows.Next() {
		var row CommanderStoreupAwardProgress
		if err := rows.Scan(&row.CommanderID, &row.StoreupID, &row.LastAwardIndex); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func SetCommanderStoreupAwardIndexTx(ctx context.Context, tx pgx.Tx, commanderID uint32, storeupID uint32, lastAwardIndex uint32) error {
	_, err := tx.Exec(ctx, `
INSERT INTO commander_storeup_award_progresses (commander_id, storeup_id, last_award_index)
VALUES ($1, $2, $3)
ON CONFLICT (commander_id, storeup_id)
DO UPDATE SET last_award_index = EXCLUDED.last_award_index
`, int64(commanderID), int64(storeupID), int64(lastAwardIndex))
	return err
}

// TryAdvanceCommanderStoreupAwardIndexTx atomically advances the storeup progress by exactly one tier.
//
// It returns (true, nil) only if the index was advanced from (awardIndex-1) -> awardIndex.
// This is used to prevent duplicate claims on concurrent requests.
func TryAdvanceCommanderStoreupAwardIndexTx(ctx context.Context, tx pgx.Tx, commanderID uint32, storeupID uint32, awardIndex uint32) (bool, error) {
	if awardIndex == 0 {
		return false, errors.New("award index must be > 0")
	}

	if awardIndex == 1 {
		res, err := tx.Exec(ctx, `
INSERT INTO commander_storeup_award_progresses (commander_id, storeup_id, last_award_index)
VALUES ($1, $2, 1)
ON CONFLICT (commander_id, storeup_id)
DO UPDATE SET last_award_index = 1
WHERE commander_storeup_award_progresses.last_award_index = 0
`, int64(commanderID), int64(storeupID))
		if err != nil {
			return false, err
		}
		return res.RowsAffected() == 1, nil
	}

	previousIndex := awardIndex - 1
	res, err := tx.Exec(ctx, `
UPDATE commander_storeup_award_progresses
SET last_award_index = $4
WHERE commander_id = $1 AND storeup_id = $2 AND last_award_index = $3
`, int64(commanderID), int64(storeupID), int64(previousIndex), int64(awardIndex))
	if err != nil {
		return false, err
	}
	return res.RowsAffected() == 1, nil
}
