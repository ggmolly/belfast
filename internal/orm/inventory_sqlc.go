package orm

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/ggmolly/belfast/internal/db"
	dbgen "github.com/ggmolly/belfast/internal/db/gen"
)

func decrementCommanderItemIfEnoughSQLC(commanderID uint32, itemID uint32, count uint32) (bool, error) {
	ctx := context.Background()
	res, err := db.DefaultStore.Queries.DecrementCommanderItemIfEnough(ctx, dbgen.DecrementCommanderItemIfEnoughParams{
		CommanderID: int64(commanderID),
		ItemID:      int64(itemID),
		Count:       int64(count),
	})
	if err != nil {
		return false, err
	}
	return rowsAffected(res) > 0, nil
}

func incrementCommanderItemSQLC(commanderID uint32, itemID uint32, amount uint32) error {
	ctx := context.Background()
	return db.DefaultStore.Queries.IncrementCommanderItem(ctx, dbgen.IncrementCommanderItemParams{
		CommanderID: int64(commanderID),
		ItemID:      int64(itemID),
		Count:       int64(amount),
	})
}

func upsertCommanderItemSetSQLC(commanderID uint32, itemID uint32, amount uint32) error {
	ctx := context.Background()
	return db.DefaultStore.Queries.UpsertCommanderItemSet(ctx, dbgen.UpsertCommanderItemSetParams{
		CommanderID: int64(commanderID),
		ItemID:      int64(itemID),
		Count:       int64(amount),
	})
}

func decrementOwnedResourceIfEnoughSQLC(commanderID uint32, resourceID uint32, count uint32) (bool, error) {
	ctx := context.Background()
	res, err := db.DefaultStore.Queries.DecrementOwnedResourceIfEnough(ctx, dbgen.DecrementOwnedResourceIfEnoughParams{
		CommanderID: int64(commanderID),
		ResourceID:  int64(resourceID),
		Amount:      int64(count),
	})
	if err != nil {
		return false, err
	}
	return rowsAffected(res) > 0, nil
}

func incrementOwnedResourceSQLC(commanderID uint32, resourceID uint32, amount uint32) error {
	ctx := context.Background()
	return db.DefaultStore.Queries.IncrementOwnedResource(ctx, dbgen.IncrementOwnedResourceParams{
		CommanderID: int64(commanderID),
		ResourceID:  int64(resourceID),
		Amount:      int64(amount),
	})
}

func upsertOwnedResourceSetSQLC(commanderID uint32, resourceID uint32, amount uint32) error {
	ctx := context.Background()
	return db.DefaultStore.Queries.UpsertOwnedResourceSet(ctx, dbgen.UpsertOwnedResourceSetParams{
		CommanderID: int64(commanderID),
		ResourceID:  int64(resourceID),
		Amount:      int64(amount),
	})
}

func rowsAffected(result any) int64 {
	// sqlc uses pgconn.CommandTag for :execresult with pgx.
	if tag, ok := result.(pgconn.CommandTag); ok {
		return tag.RowsAffected()
	}
	return 0
}
