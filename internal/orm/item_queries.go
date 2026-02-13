package orm

import (
	"context"

	"github.com/ggmolly/belfast/internal/db"
)

func ListAllItems() ([]Item, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT id, name, rarity, shop_id, type, virtual_type
FROM items
ORDER BY id ASC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Item, 0)
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Rarity, &item.ShopID, &item.Type, &item.VirtualType); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
