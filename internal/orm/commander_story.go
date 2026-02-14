package orm

import (
	"context"
	"time"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

type CommanderStory struct {
	CommanderID uint32    `gorm:"primaryKey;autoIncrement:false"`
	StoryID     uint32    `gorm:"primaryKey;autoIncrement:false"`
	CreatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
}

func ListCommanderStoryIDs(commanderID uint32) ([]uint32, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListCommanderStories(ctx, int64(commanderID))
	if err != nil {
		return nil, err
	}
	ids := make([]uint32, 0, len(rows))
	for _, id := range rows {
		ids = append(ids, uint32(id))
	}
	return ids, nil
}

func AddCommanderStory(commanderID uint32, storyID uint32) error {
	ctx := context.Background()
	return db.DefaultStore.Queries.CreateCommanderStory(ctx, gen.CreateCommanderStoryParams{CommanderID: int64(commanderID), StoryID: int64(storyID)})
}

func DeleteCommanderStory(commanderID uint32, storyID uint32) error {
	ctx := context.Background()
	res, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM commander_stories
WHERE commander_id = $1
  AND story_id = $2
`, int64(commanderID), int64(storyID))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}
