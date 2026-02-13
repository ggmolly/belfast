package orm

import (
	"context"

	"github.com/ggmolly/belfast/internal/db"
)

func GetMailboxCounts(commanderID uint32) (uint32, uint32, error) {
	ctx := context.Background()
	var total int64
	err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT COUNT(*)::bigint
FROM mails
WHERE receiver_id = $1 AND is_archived = false
`, int64(commanderID)).Scan(&total)
	if err != nil {
		return 0, 0, err
	}

	var unread int64
	err = db.DefaultStore.Pool.QueryRow(ctx, `
SELECT COUNT(*)::bigint
FROM mails
WHERE receiver_id = $1 AND is_archived = false AND read = false
`, int64(commanderID)).Scan(&unread)
	if err != nil {
		return 0, 0, err
	}

	return uint32(total), uint32(unread), nil
}
