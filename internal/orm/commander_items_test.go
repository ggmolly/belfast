package orm

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
)

var commanderItemTestOnce sync.Once

func initCommanderItemTestDB(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	commanderItemTestOnce.Do(func() {
		InitDatabase()
	})
}

func clearTable(t *testing.T, model any) {
	t.Helper()
	tableName := testTableName(model)
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), fmt.Sprintf("DELETE FROM %s", QualifiedTable(tableName))); err != nil {
		t.Fatalf("clear table: %v", err)
	}
}

func testTableName(model any) string {
	t := reflect.TypeOf(model)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	name := t.Name()
	if name == "" {
		panic("model type has no name")
	}
	return pluralizeSnakeCase(name)
}

func pluralizeSnakeCase(name string) string {
	var b strings.Builder
	for i, r := range name {
		if i > 0 {
			prev := rune(name[i-1])
			if isUpperASCII(r) && (!isUpperASCII(prev) || (i+1 < len(name) && !isUpperASCII(rune(name[i+1])))) {
				b.WriteByte('_')
			}
		}
		if 'A' <= r && r <= 'Z' {
			b.WriteByte(byte(r + ('a' - 'A')))
			continue
		}
		b.WriteRune(r)
	}
	s := b.String()
	if strings.HasSuffix(s, "y") && len(s) > 1 {
		last := s[len(s)-2]
		if !strings.ContainsRune("aeiou", rune(last)) {
			return s[:len(s)-1] + "ies"
		}
	}
	if strings.HasSuffix(s, "s") || strings.HasSuffix(s, "x") || strings.HasSuffix(s, "z") || strings.HasSuffix(s, "ch") || strings.HasSuffix(s, "sh") {
		return s + "es"
	}
	return s + "s"
}

func isUpperASCII(r rune) bool {
	return 'A' <= r && r <= 'Z'
}

func TestCommanderAddItemUpdatesExistingRow(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &CommanderItem{})
	clearTable(t, &Item{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 1001, AccountID: 1, Name: "Tester"}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commanders (commander_id, account_id, name) VALUES ($1, $2, $3)`, int64(commander.CommanderID), int64(commander.AccountID), commander.Name); err != nil {
		t.Fatalf("create commander: %v", err)
	}
	item := Item{ID: 30041, Name: "Test Item", Rarity: 1, ShopID: -2, Type: 1, VirtualType: 0}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO items (id, name, rarity, shop_id, type, virtual_type) VALUES ($1, $2, $3, $4, $5, $6)`, int64(item.ID), item.Name, item.Rarity, item.ShopID, item.Type, item.VirtualType); err != nil {
		t.Fatalf("create item: %v", err)
	}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)`, int64(commander.CommanderID), int64(30041), int64(1)); err != nil {
		t.Fatalf("seed commander item: %v", err)
	}
	commander.CommanderItemsMap = make(map[uint32]*CommanderItem)
	commander.Items = []CommanderItem{}

	if err := commander.AddItem(30041, 1); err != nil {
		t.Fatalf("add item: %v", err)
	}
	var stored CommanderItem
	if err := db.DefaultStore.Pool.QueryRow(context.Background(), `SELECT commander_id, item_id, count FROM commander_items WHERE commander_id = $1 AND item_id = $2`, int64(commander.CommanderID), int64(30041)).Scan(&stored.CommanderID, &stored.ItemID, &stored.Count); err != nil {
		t.Fatalf("load commander item: %v", err)
	}
	if stored.Count != 2 {
		t.Fatalf("expected count 2, got %d", stored.Count)
	}
	if commander.CommanderItemsMap[30041] == nil || commander.CommanderItemsMap[30041].Count != 2 {
		t.Fatalf("expected map count 2, got %+v", commander.CommanderItemsMap[30041])
	}
}

func TestCommanderAddResourceUpdatesExistingRow(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &OwnedResource{})
	clearTable(t, &Resource{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 1002, AccountID: 1, Name: "Tester"}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commanders (commander_id, account_id, name) VALUES ($1, $2, $3)`, int64(commander.CommanderID), int64(commander.AccountID), commander.Name); err != nil {
		t.Fatalf("create commander: %v", err)
	}
	resource := Resource{ID: 2, Name: "Oil"}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO resources (id, item_id, name) VALUES ($1, $2, $3)`, int64(resource.ID), int64(resource.ItemID), resource.Name); err != nil {
		t.Fatalf("create resource: %v", err)
	}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)`, int64(commander.CommanderID), int64(2), int64(5)); err != nil {
		t.Fatalf("seed owned resource: %v", err)
	}
	commander.OwnedResourcesMap = make(map[uint32]*OwnedResource)
	commander.OwnedResources = []OwnedResource{}

	if err := commander.AddResource(2, 3); err != nil {
		t.Fatalf("add resource: %v", err)
	}
	var stored OwnedResource
	if err := db.DefaultStore.Pool.QueryRow(context.Background(), `SELECT commander_id, resource_id, amount FROM owned_resources WHERE commander_id = $1 AND resource_id = $2`, int64(commander.CommanderID), int64(2)).Scan(&stored.CommanderID, &stored.ResourceID, &stored.Amount); err != nil {
		t.Fatalf("load resource: %v", err)
	}
	if stored.Amount != 8 {
		t.Fatalf("expected amount 8, got %d", stored.Amount)
	}
	if commander.OwnedResourcesMap[2] == nil || commander.OwnedResourcesMap[2].Amount != 8 {
		t.Fatalf("expected map amount 8, got %+v", commander.OwnedResourcesMap[2])
	}
}
