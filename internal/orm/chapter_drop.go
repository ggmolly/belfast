package orm

import (
	"context"

	"github.com/ggmolly/belfast/internal/db"
)

// ChapterDrop tracks unique ship drops for a commander in a given chapter.
// Rows are inserted as ships are obtained; duplicates are ignored.
type ChapterDrop struct {
	CommanderID uint32 `gorm:"primaryKey;autoIncrement:false"`
	ChapterID   uint32 `gorm:"primaryKey;autoIncrement:false"`
	ShipID      uint32 `gorm:"primaryKey;autoIncrement:false"`
}

func GetChapterDrops(commanderID uint32, chapterID uint32) ([]ChapterDrop, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT commander_id, chapter_id, ship_id
FROM chapter_drops
WHERE commander_id = $1 AND chapter_id = $2
ORDER BY ship_id ASC
`, int64(commanderID), int64(chapterID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	drops := make([]ChapterDrop, 0)
	for rows.Next() {
		var row ChapterDrop
		if err := rows.Scan(&row.CommanderID, &row.ChapterID, &row.ShipID); err != nil {
			return nil, err
		}
		drops = append(drops, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return drops, nil
}

func AddChapterDrop(drop *ChapterDrop) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO chapter_drops (commander_id, chapter_id, ship_id)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING
`, int64(drop.CommanderID), int64(drop.ChapterID), int64(drop.ShipID))
	return err
}
