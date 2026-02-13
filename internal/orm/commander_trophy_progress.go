package orm

import (
	"context"

	"github.com/ggmolly/belfast/internal/db"
)

// CommanderTrophyProgress tracks a commander's trophy/medal progress and claim timestamp.
// Timestamp == 0 means unclaimed.
type CommanderTrophyProgress struct {
	CommanderID uint32 `gorm:"primaryKey;autoIncrement:false"`
	TrophyID    uint32 `gorm:"primaryKey;autoIncrement:false"`
	Progress    uint32 `gorm:"not null;default:0"`
	Timestamp   uint32 `gorm:"not null;default:0"`
}

func GetCommanderTrophyProgress(commanderID uint32, trophyID uint32) (*CommanderTrophyProgress, error) {
	ctx := context.Background()
	row := CommanderTrophyProgress{}
	err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT commander_id, trophy_id, progress, timestamp
FROM commander_trophy_progresses
WHERE commander_id = $1 AND trophy_id = $2
`, int64(commanderID), int64(trophyID)).Scan(&row.CommanderID, &row.TrophyID, &row.Progress, &row.Timestamp)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func GetOrCreateCommanderTrophyProgress(commanderID uint32, trophyID uint32, progress uint32) (*CommanderTrophyProgress, bool, error) {
	row, err := GetCommanderTrophyProgress(commanderID, trophyID)
	if err == nil {
		return row, false, nil
	}
	if !db.IsNotFound(err) {
		return nil, false, err
	}
	row = &CommanderTrophyProgress{
		CommanderID: commanderID,
		TrophyID:    trophyID,
		Progress:    progress,
		Timestamp:   0,
	}
	if err := UpdateCommanderTrophyProgress(row); err != nil {
		return nil, false, err
	}
	return row, true, nil
}

func UpdateCommanderTrophyProgress(row *CommanderTrophyProgress) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO commander_trophy_progresses (commander_id, trophy_id, progress, timestamp)
VALUES ($1, $2, $3, $4)
ON CONFLICT (commander_id, trophy_id)
DO UPDATE SET progress = EXCLUDED.progress, timestamp = EXCLUDED.timestamp
`, int64(row.CommanderID), int64(row.TrophyID), int64(row.Progress), int64(row.Timestamp))
	return err
}

func ClaimCommanderTrophyProgress(commanderID uint32, trophyID uint32, timestamp uint32) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE commander_trophy_progresses
SET timestamp = $3
WHERE commander_id = $1 AND trophy_id = $2
`, int64(commanderID), int64(trophyID), int64(timestamp))
	return err
}
