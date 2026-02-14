package orm

import (
	"context"

	"github.com/ggmolly/belfast/internal/db"
)

func ListAllResources() ([]Resource, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT id, item_id, name
FROM resources
ORDER BY id ASC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resources := make([]Resource, 0)
	for rows.Next() {
		var resource Resource
		if err := rows.Scan(&resource.ID, &resource.ItemID, &resource.Name); err != nil {
			return nil, err
		}
		resources = append(resources, resource)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return resources, nil
}

func DeleteOwnedResource(commanderID uint32, resourceID uint32) error {
	ctx := context.Background()
	res, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM owned_resources
WHERE commander_id = $1
  AND resource_id = $2
`, int64(commanderID), int64(resourceID))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}
