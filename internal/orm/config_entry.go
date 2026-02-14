package orm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

type ConfigEntry struct {
	ID       uint64          `json:"id"`
	Category string          `json:"category"`
	Key      string          `json:"key"`
	Data     json.RawMessage `json:"data"`
}

func ListConfigEntries(category string) ([]ConfigEntry, error) {
	if db.DefaultStore == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListConfigEntriesByCategory(ctx, category)
	if err != nil {
		return nil, err
	}
	entries := make([]ConfigEntry, 0, len(rows))
	for _, r := range rows {
		entries = append(entries, ConfigEntry{ID: uint64(r.ID), Category: r.Category, Key: r.Key, Data: r.Data})
	}
	return entries, nil
}

func GetConfigEntry(category string, key string) (*ConfigEntry, error) {
	if db.DefaultStore == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetConfigEntry(ctx, gen.GetConfigEntryParams{Category: category, Key: key})
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	entry := ConfigEntry{ID: uint64(row.ID), Category: row.Category, Key: row.Key, Data: row.Data}
	return &entry, nil
}

func UpsertConfigEntry(category string, key string, data json.RawMessage) error {
	if db.DefaultStore == nil {
		return fmt.Errorf("database is not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.Queries.UpsertConfigEntry(ctx, gen.UpsertConfigEntryParams{Category: category, Key: key, Data: data})
}
