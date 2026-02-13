package orm

import (
	"context"

	"github.com/ggmolly/belfast/internal/db"
)

func ListLikesByCommander(commanderID uint32) ([]Like, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT group_id, liker_id, created_at
FROM likes
WHERE liker_id = $1
ORDER BY group_id ASC
`, int64(commanderID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	likes := make([]Like, 0)
	for rows.Next() {
		var like Like
		if err := rows.Scan(&like.GroupID, &like.LikerID, &like.CreatedAt); err != nil {
			return nil, err
		}
		likes = append(likes, like)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return likes, nil
}

func DeleteLike(commanderID uint32, groupID uint32) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM likes
WHERE liker_id = $1
  AND group_id = $2
`, int64(commanderID), int64(groupID))
	return err
}
