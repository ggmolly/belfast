package orm

import (
	"context"
	"time"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

type CommanderSoundStory struct {
	CommanderID  uint32    `gorm:"primaryKey;autoIncrement:false"`
	SoundStoryID uint32    `gorm:"primaryKey;autoIncrement:false"`
	CreatedAt    time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
}

func ListCommanderSoundStoryIDs(commanderID uint32) ([]uint32, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListCommanderSoundStories(ctx, int64(commanderID))
	if err != nil {
		return nil, err
	}
	ids := make([]uint32, 0, len(rows))
	for _, id := range rows {
		ids = append(ids, uint32(id))
	}
	return ids, nil
}

func IsCommanderSoundStoryUnlockedTx(q *gen.Queries, commanderID uint32, soundStoryID uint32) (bool, error) {
	ctx := context.Background()
	exists, err := q.HasCommanderSoundStory(ctx, gen.HasCommanderSoundStoryParams{CommanderID: int64(commanderID), StoryID: int64(soundStoryID)})
	return exists, err
}

func UnlockCommanderSoundStoryTx(q *gen.Queries, commanderID uint32, soundStoryID uint32) error {
	ctx := context.Background()
	return q.CreateCommanderSoundStory(ctx, gen.CreateCommanderSoundStoryParams{CommanderID: int64(commanderID), StoryID: int64(soundStoryID)})
}
