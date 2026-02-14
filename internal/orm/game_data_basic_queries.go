package orm

import (
	"context"

	"github.com/ggmolly/belfast/internal/db"
)

func ListShipTypes(offset int, limit int) ([]ShipType, int64, error) {
	ctx := context.Background()
	var total int64
	if err := db.DefaultStore.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM ship_types`).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset, limit, unlimited := normalizePagination(offset, limit)
	query := `
SELECT id, name
FROM ship_types
ORDER BY id ASC
OFFSET $1
`
	args := []any{int64(offset)}
	if !unlimited {
		query += `LIMIT $2`
		args = append(args, int64(limit))
	}
	rows, err := db.DefaultStore.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	shipTypes := make([]ShipType, 0)
	for rows.Next() {
		var shipType ShipType
		if err := rows.Scan(&shipType.ID, &shipType.Name); err != nil {
			return nil, 0, err
		}
		shipTypes = append(shipTypes, shipType)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return shipTypes, total, nil
}

func GetShipTypeByID(id uint32) (*ShipType, error) {
	ctx := context.Background()
	var shipType ShipType
	err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT id, name
FROM ship_types
WHERE id = $1
`, int64(id)).Scan(&shipType.ID, &shipType.Name)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &shipType, nil
}

func CreateShipType(shipType *ShipType) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO ship_types (id, name)
VALUES ($1, $2)
`, int64(shipType.ID), shipType.Name)
	return err
}

func UpdateShipType(shipType *ShipType) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE ship_types
SET name = $2
WHERE id = $1
`, int64(shipType.ID), shipType.Name)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func DeleteShipType(id uint32) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM ship_types WHERE id = $1`, int64(id))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func ListRarities(offset int, limit int) ([]Rarity, int64, error) {
	ctx := context.Background()
	var total int64
	if err := db.DefaultStore.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM rarities`).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset, limit, unlimited := normalizePagination(offset, limit)
	query := `
SELECT id, name
FROM rarities
ORDER BY id ASC
OFFSET $1
`
	args := []any{int64(offset)}
	if !unlimited {
		query += `LIMIT $2`
		args = append(args, int64(limit))
	}
	rows, err := db.DefaultStore.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	rarities := make([]Rarity, 0)
	for rows.Next() {
		var rarity Rarity
		if err := rows.Scan(&rarity.ID, &rarity.Name); err != nil {
			return nil, 0, err
		}
		rarities = append(rarities, rarity)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return rarities, total, nil
}

func GetRarityByID(id uint32) (*Rarity, error) {
	ctx := context.Background()
	var rarity Rarity
	err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT id, name
FROM rarities
WHERE id = $1
`, int64(id)).Scan(&rarity.ID, &rarity.Name)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &rarity, nil
}

func CreateRarity(rarity *Rarity) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO rarities (id, name)
VALUES ($1, $2)
`, int64(rarity.ID), rarity.Name)
	return err
}

func UpdateRarity(rarity *Rarity) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE rarities
SET name = $2
WHERE id = $1
`, int64(rarity.ID), rarity.Name)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func DeleteRarity(id uint32) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM rarities WHERE id = $1`, int64(id))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func ListItemsPage(offset int, limit int) ([]Item, int64, error) {
	ctx := context.Background()
	var total int64
	if err := db.DefaultStore.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM items`).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset, limit, unlimited := normalizePagination(offset, limit)
	query := `
SELECT id, name, rarity, shop_id, type, virtual_type
FROM items
ORDER BY id ASC
OFFSET $1
`
	args := []any{int64(offset)}
	if !unlimited {
		query += `LIMIT $2`
		args = append(args, int64(limit))
	}
	rows, err := db.DefaultStore.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]Item, 0)
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Rarity, &item.ShopID, &item.Type, &item.VirtualType); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func GetItemByID(id uint32) (*Item, error) {
	ctx := context.Background()
	var item Item
	err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT id, name, rarity, shop_id, type, virtual_type
FROM items
WHERE id = $1
`, int64(id)).Scan(&item.ID, &item.Name, &item.Rarity, &item.ShopID, &item.Type, &item.VirtualType)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func CreateItemRecord(item *Item) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO items (id, name, rarity, shop_id, type, virtual_type)
VALUES ($1, $2, $3, $4, $5, $6)
`, int64(item.ID), item.Name, int64(item.Rarity), int64(item.ShopID), int64(item.Type), int64(item.VirtualType))
	return err
}

func UpdateItemRecord(item *Item) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE items
SET name = $2,
	rarity = $3,
	shop_id = $4,
	type = $5,
	virtual_type = $6
WHERE id = $1
`, int64(item.ID), item.Name, int64(item.Rarity), int64(item.ShopID), int64(item.Type), int64(item.VirtualType))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func DeleteItemRecord(id uint32) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM items WHERE id = $1`, int64(id))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func ListResourcesPage(offset int, limit int) ([]Resource, int64, error) {
	ctx := context.Background()
	var total int64
	if err := db.DefaultStore.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM resources`).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset, limit, unlimited := normalizePagination(offset, limit)
	query := `
SELECT id, item_id, name
FROM resources
ORDER BY id ASC
OFFSET $1
`
	args := []any{int64(offset)}
	if !unlimited {
		query += `LIMIT $2`
		args = append(args, int64(limit))
	}
	rows, err := db.DefaultStore.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	resources := make([]Resource, 0)
	for rows.Next() {
		var resource Resource
		if err := rows.Scan(&resource.ID, &resource.ItemID, &resource.Name); err != nil {
			return nil, 0, err
		}
		resources = append(resources, resource)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return resources, total, nil
}

func GetResourceByID(id uint32) (*Resource, error) {
	ctx := context.Background()
	var resource Resource
	err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT id, item_id, name
FROM resources
WHERE id = $1
`, int64(id)).Scan(&resource.ID, &resource.ItemID, &resource.Name)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

func CreateResourceRecord(resource *Resource) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO resources (id, item_id, name)
VALUES ($1, $2, $3)
`, int64(resource.ID), int64(resource.ItemID), resource.Name)
	return err
}

func UpdateResourceRecord(resource *Resource) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE resources
SET item_id = $2,
	name = $3
WHERE id = $1
`, int64(resource.ID), int64(resource.ItemID), resource.Name)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func DeleteResourceRecord(id uint32) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM resources WHERE id = $1`, int64(id))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}
