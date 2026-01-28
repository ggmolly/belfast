package orm

import (
	"sync"
	"testing"

	"gorm.io/gorm"
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
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(model).Error; err != nil {
		t.Fatalf("clear table: %v", err)
	}
}

func TestCommanderAddItemUpdatesExistingRow(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &CommanderItem{})
	clearTable(t, &Item{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 1001, AccountID: 1, Name: "Tester"}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}
	item := Item{ID: 30041, Name: "Test Item", Rarity: 1, ShopID: -2, Type: 1, VirtualType: 0}
	if err := GormDB.Create(&item).Error; err != nil {
		t.Fatalf("create item: %v", err)
	}
	if err := GormDB.Create(&CommanderItem{CommanderID: commander.CommanderID, ItemID: 30041, Count: 1}).Error; err != nil {
		t.Fatalf("seed commander item: %v", err)
	}
	commander.CommanderItemsMap = make(map[uint32]*CommanderItem)
	commander.Items = []CommanderItem{}

	if err := commander.AddItem(30041, 1); err != nil {
		t.Fatalf("add item: %v", err)
	}
	var stored CommanderItem
	if err := GormDB.First(&stored, "commander_id = ? AND item_id = ?", commander.CommanderID, 30041).Error; err != nil {
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
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}
	resource := Resource{ID: 2, Name: "Oil"}
	if err := GormDB.Create(&resource).Error; err != nil {
		t.Fatalf("create resource: %v", err)
	}
	if err := GormDB.Create(&OwnedResource{CommanderID: commander.CommanderID, ResourceID: 2, Amount: 5}).Error; err != nil {
		t.Fatalf("seed owned resource: %v", err)
	}
	commander.OwnedResourcesMap = make(map[uint32]*OwnedResource)
	commander.OwnedResources = []OwnedResource{}

	if err := commander.AddResource(2, 3); err != nil {
		t.Fatalf("add resource: %v", err)
	}
	var stored OwnedResource
	if err := GormDB.First(&stored, "commander_id = ? AND resource_id = ?", commander.CommanderID, 2).Error; err != nil {
		t.Fatalf("load resource: %v", err)
	}
	if stored.Amount != 8 {
		t.Fatalf("expected amount 8, got %d", stored.Amount)
	}
	if commander.OwnedResourcesMap[2] == nil || commander.OwnedResourcesMap[2].Amount != 8 {
		t.Fatalf("expected map amount 8, got %+v", commander.OwnedResourcesMap[2])
	}
}
