package orm

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/db"
)

func TestBuildCRUDAndRetrieve(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Build{})
	clearTable(t, &Ship{})
	clearTable(t, &Commander{})
	clearTable(t, &Rarity{})

	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO rarities (id, name) VALUES ($1, $2)`, int64(2), "Common"); err != nil {
		t.Fatalf("seed rarity: %v", err)
	}
	ship := Ship{TemplateID: 1001, Name: "Test", EnglishName: "Test", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	if err := ship.Create(); err != nil {
		t.Fatalf("seed ship: %v", err)
	}
	commander := Commander{CommanderID: 10, AccountID: 10, Name: "Builder"}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commanders (commander_id, account_id, name) VALUES ($1, $2, $3)`, int64(commander.CommanderID), int64(commander.AccountID), commander.Name); err != nil {
		t.Fatalf("seed commander: %v", err)
	}

	build := Build{ID: 1, BuilderID: commander.CommanderID, ShipID: ship.TemplateID, PoolID: 3, FinishesAt: time.Now()}
	if err := build.Create(); err != nil {
		t.Fatalf("create build: %v", err)
	}
	build.PoolID = 4
	if err := build.Update(); err != nil {
		t.Fatalf("update build: %v", err)
	}
	loaded := Build{ID: 1}
	if err := loaded.Retrieve(false); err != nil {
		t.Fatalf("retrieve build: %v", err)
	}
	if loaded.PoolID != 4 {
		t.Fatalf("expected pool id 4, got %d", loaded.PoolID)
	}
	if err := loaded.Retrieve(true); err != nil {
		t.Fatalf("retrieve build greedy: %v", err)
	}
	if _, err := GetBuildByID(1); err != nil {
		t.Fatalf("get build by id: %v", err)
	}
	if _, err := GetBuildByID(999); !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("expected not found for missing build, got %v", err)
	}
	if err := loaded.Delete(); err != nil {
		t.Fatalf("delete build: %v", err)
	}
}

func TestBuildConsume(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Build{})
	clearTable(t, &OwnedShip{})
	clearTable(t, &Ship{})
	clearTable(t, &Commander{})

	ship := Ship{TemplateID: 2001, Name: "Ship", EnglishName: "Ship", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	if err := ship.Create(); err != nil {
		t.Fatalf("seed ship: %v", err)
	}
	commander := Commander{CommanderID: 20, AccountID: 20, Name: "Consumer"}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commanders (commander_id, account_id, name) VALUES ($1, $2, $3)`, int64(commander.CommanderID), int64(commander.AccountID), commander.Name); err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	commander.OwnedShipsMap = make(map[uint32]*OwnedShip)
	commander.BuildsMap = make(map[uint32]*Build)

	build := Build{ID: 2, BuilderID: commander.CommanderID, ShipID: ship.TemplateID, PoolID: 1, FinishesAt: time.Now()}
	if err := build.Create(); err != nil {
		t.Fatalf("create build: %v", err)
	}
	commander.Builds = []Build{build}
	commander.BuildsMap[build.ID] = &commander.Builds[0]

	owned, err := build.Consume(ship.TemplateID, &commander)
	if err != nil {
		t.Fatalf("consume build: %v", err)
	}
	if owned == nil || owned.ShipID != ship.TemplateID {
		t.Fatalf("expected owned ship created")
	}
	if len(commander.Builds) != 0 {
		t.Fatalf("expected build removed from commander")
	}
	var count int64
	if err := db.DefaultStore.Pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM builds WHERE id = $1`, int64(build.ID)).Scan(&count); err != nil {
		t.Fatalf("count builds: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected build deleted")
	}
}

func TestBuildQuickFinish(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Build{})
	clearTable(t, &CommanderItem{})
	clearTable(t, &Item{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 30, AccountID: 30, Name: "Finisher"}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commanders (commander_id, account_id, name) VALUES ($1, $2, $3)`, int64(commander.CommanderID), int64(commander.AccountID), commander.Name); err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	item := Item{ID: 15003, Name: "Quick Finisher", Rarity: 1, ShopID: -2, Type: 1, VirtualType: 0}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO items (id, name, rarity, shop_id, type, virtual_type) VALUES ($1, $2, $3, $4, $5, $6)`, int64(item.ID), item.Name, item.Rarity, item.ShopID, item.Type, item.VirtualType); err != nil {
		t.Fatalf("seed item: %v", err)
	}
	cmdItem := CommanderItem{CommanderID: commander.CommanderID, ItemID: 15003, Count: 1}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commander_items (commander_id, item_id, count) VALUES ($1, $2, $3)`, int64(cmdItem.CommanderID), int64(cmdItem.ItemID), int64(cmdItem.Count)); err != nil {
		t.Fatalf("seed commander item: %v", err)
	}
	commander.CommanderItemsMap = map[uint32]*CommanderItem{15003: &cmdItem}

	build := Build{ID: 3, BuilderID: commander.CommanderID, ShipID: 1, PoolID: 1, FinishesAt: time.Now().Add(time.Hour)}
	if err := build.Create(); err != nil {
		t.Fatalf("create build: %v", err)
	}
	if err := build.QuickFinish(&commander); err != nil {
		t.Fatalf("quick finish: %v", err)
	}
	stored, err := GetBuildByID(build.ID)
	if err != nil {
		t.Fatalf("load build: %v", err)
	}
	if stored.FinishesAt.After(time.Now()) {
		t.Fatalf("expected finish time in past")
	}
	var storedItem CommanderItem
	if err := db.DefaultStore.Pool.QueryRow(context.Background(), `SELECT commander_id, item_id, count FROM commander_items WHERE commander_id = $1 AND item_id = $2`, int64(commander.CommanderID), int64(15003)).Scan(&storedItem.CommanderID, &storedItem.ItemID, &storedItem.Count); err != nil {
		t.Fatalf("load commander item: %v", err)
	}
	if storedItem.Count != 0 {
		t.Fatalf("expected quick finisher consumed")
	}

	commander.CommanderItemsMap = map[uint32]*CommanderItem{}
	if err := build.QuickFinish(&commander); !errors.Is(err, ErrorNotEnoughQuickFinishers) {
		t.Fatalf("expected not enough quick finishers, got %v", err)
	}
}
