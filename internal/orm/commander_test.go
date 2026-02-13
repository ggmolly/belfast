package orm

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ggmolly/belfast/internal/db"
)

var fakeCommander Commander

var (
	fakeResources         []Resource
	fakeItems             []Item
	commanderTestInitOnce sync.Once
)

func seedDb() {
	commanderTestInitOnce.Do(func() {
		os.Setenv("MODE", "test")
		InitDatabase()
	})

	err := db.DefaultStore.WithPGXTx(context.Background(), func(tx pgx.Tx) error {
		for _, r := range fakeResources {
			if _, err := tx.Exec(context.Background(), `
INSERT INTO resources (id, item_id, name)
VALUES ($1, $2, $3)
ON CONFLICT (id) DO UPDATE SET item_id = EXCLUDED.item_id, name = EXCLUDED.name
`, int64(r.ID), int64(r.ItemID), r.Name); err != nil {
				return err
			}
		}
		for _, i := range fakeItems {
			if _, err := tx.Exec(context.Background(), `
INSERT INTO items (id, name, rarity, shop_id, type, virtual_type)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (id) DO UPDATE SET
	name = EXCLUDED.name,
	rarity = EXCLUDED.rarity,
	shop_id = EXCLUDED.shop_id,
	type = EXCLUDED.type,
	virtual_type = EXCLUDED.virtual_type
`, int64(i.ID), i.Name, i.Rarity, i.ShopID, i.Type, i.VirtualType); err != nil {
				return err
			}
		}

		// Reset commander-owned rows so these tests don't depend on package test order.
		if _, err := tx.Exec(context.Background(), `DELETE FROM owned_resources WHERE commander_id = $1`, int64(fakeCommander.CommanderID)); err != nil {
			return err
		}
		if _, err := tx.Exec(context.Background(), `DELETE FROM commander_items WHERE commander_id = $1`, int64(fakeCommander.CommanderID)); err != nil {
			return err
		}
		if _, err := tx.Exec(context.Background(), `DELETE FROM commander_misc_items WHERE commander_id = $1`, int64(fakeCommander.CommanderID)); err != nil {
			return err
		}

		if _, err := tx.Exec(context.Background(), `
INSERT INTO commanders (commander_id, account_id, level, exp, name, last_login, guide_index, new_guide_index, name_change_cooldown, room_id, exchange_count, draw_count1, draw_count10, support_requisition_count, support_requisition_month, collect_attack_count, acc_pay_lv, living_area_cover_id, selected_icon_frame_id, selected_chat_frame_id, selected_battle_ui_id, display_icon_id, display_skin_id, display_icon_theme_id, manifesto, dorm_name, random_ship_mode, random_flag_ship_enabled, deleted_at)
VALUES ($1, $2, 1, 0, $3, now(), 0, 0, '1970-01-01 00:00:00+00', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, '', '', 0, false, NULL)
ON CONFLICT (commander_id) DO UPDATE SET
	account_id = EXCLUDED.account_id,
	name = EXCLUDED.name,
	deleted_at = NULL
`, int64(fakeCommander.CommanderID), int64(fakeCommander.AccountID), fakeCommander.Name); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	fakeCommander.OwnedResources = nil
	fakeCommander.Items = nil
	fakeCommander.MiscItems = nil
	fakeCommander.OwnedResourcesMap = make(map[uint32]*OwnedResource)
	fakeCommander.CommanderItemsMap = make(map[uint32]*CommanderItem)
	fakeCommander.MiscItemsMap = make(map[uint32]*CommanderMiscItem)

	// Fake resources
	fakeResourcesCnt := []uint32{100, 30}
	for i := 0; i < len(fakeResources); i++ {
		resource := &OwnedResource{
			ResourceID:  fakeResources[i].ID,
			Amount:      fakeResourcesCnt[i],
			CommanderID: fakeCommander.CommanderID,
		}
		if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO owned_resources (resource_id, amount, commander_id) VALUES ($1, $2, $3)`, int64(resource.ResourceID), int64(resource.Amount), int64(resource.CommanderID)); err != nil {
			panic(err)
		}
		fakeCommander.OwnedResourcesMap[fakeResources[i].ID] = resource
	}

	// Fake items
	fakeItemsCnt := []uint32{5, 50, 3}
	for i := 0; i < len(fakeItems); i++ {
		item := &CommanderItem{
			ItemID:      fakeItems[i].ID,
			Count:       fakeItemsCnt[i],
			CommanderID: fakeCommander.CommanderID,
		}
		if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commander_items (item_id, count, commander_id) VALUES ($1, $2, $3)`, int64(item.ItemID), int64(item.Count), int64(item.CommanderID)); err != nil {
			panic(err)
		}
		fakeCommander.CommanderItemsMap[fakeItems[i].ID] = item
	}
}

func init() {
	fakeCommander.AccountID = 1
	fakeCommander.CommanderID = 1
	fakeCommander.Name = "Fake Commander"
	fakeResources = []Resource{
		{ID: 1, Name: "Gold"},
		{ID: 2, Name: "Fake resource"},
	}
	fakeItems = []Item{
		{ID: 20001, Name: "Wisdom Cube"},
		{ID: 45, Name: "Fake Item"},
		{ID: 60, Name: "Fake Item 2"},
	}

}

// Tests the behavior of orm.Commander.HasEnoughGold
func TestEnoughGold(t *testing.T) {
	seedDb()
	if !fakeCommander.HasEnoughGold(100) {
		t.Errorf("Expected enough gold, has %d, need %d", fakeCommander.OwnedResourcesMap[1].Amount, 100)
	}
	if !fakeCommander.HasEnoughGold(50) {
		t.Errorf("Expected enough gold, has %d, need %d", fakeCommander.OwnedResourcesMap[1].Amount, 50)
	}
	if !fakeCommander.HasEnoughGold(0) {
		t.Errorf("Expected enough gold, has %d, need %d", fakeCommander.OwnedResourcesMap[1].Amount, 0)
	}
	if fakeCommander.HasEnoughGold(1000) {
		t.Errorf("Expected not enough gold, has %d, need %d", fakeCommander.OwnedResourcesMap[1].Amount, 1000)
	}
}

// Tests the behavior of orm.Commander.HasEnoughCube
func TestEnoughCube(t *testing.T) {
	seedDb()
	if fakeCommander.HasEnoughCube(10) {
		t.Errorf("Expected not enough cube, has %d, need %d", fakeCommander.OwnedResourcesMap[2].Amount, 10)
	}
	if !fakeCommander.HasEnoughCube(5) {
		t.Errorf("Expected enough cube, has %d, need %d", fakeCommander.OwnedResourcesMap[2].Amount, 5)
	}
	if !fakeCommander.HasEnoughCube(0) {
		t.Errorf("Expected enough cube, has %d, need %d", fakeCommander.OwnedResourcesMap[2].Amount, 0)
	}
	if fakeCommander.HasEnoughCube(1000) {
		t.Errorf("Expected not enough cube, has %d, need %d", fakeCommander.OwnedResourcesMap[2].Amount, 1000)
	}
}

// Tests the behavior of orm.Commander.HasEnoughResource
func TestEnoughResource(t *testing.T) {
	seedDb()
	if !fakeCommander.HasEnoughResource(2, 1) {
		t.Errorf("Expected enough resource, has %d, need %d", fakeCommander.OwnedResourcesMap[2].Amount, 1)
	}
	if !fakeCommander.HasEnoughResource(2, 30) {
		t.Errorf("Expected enough resource, has %d, need %d", fakeCommander.OwnedResourcesMap[2].Amount, 30)
	}
	if !fakeCommander.HasEnoughResource(2, 0) {
		t.Errorf("Expected enough resource, has %d, need %d", fakeCommander.OwnedResourcesMap[2].Amount, 0)
	}
	if fakeCommander.HasEnoughResource(2, 1000) {
		t.Errorf("Expected not enough resource, has %d, need %d", fakeCommander.OwnedResourcesMap[2].Amount, 1000)
	}
	if fakeCommander.HasEnoughResource(3, 1) { // Resource not owned
		t.Errorf("Expected not enough resource, has -, need %d", 1)
	}
}

// Tests the behavior of orm.Commander.HasEnoughItem
func TestEnoughItem(t *testing.T) {
	seedDb()
	if !fakeCommander.HasEnoughItem(20001, 5) {
		t.Errorf("Expected enough item, has %d, need %d", fakeCommander.CommanderItemsMap[20001].Count, 5)
	}
	if !fakeCommander.HasEnoughItem(20001, 0) {
		t.Errorf("Expected enough item, has %d, need %d", fakeCommander.CommanderItemsMap[20001].Count, 0)
	}
	if fakeCommander.HasEnoughItem(20001, 6) {
		t.Errorf("Expected not enough item, has %d, need %d", fakeCommander.CommanderItemsMap[20001].Count, 6)
	}
	if fakeCommander.HasEnoughItem(20002, 1) { // Item not owned
		t.Errorf("Expected not enough item, has -, need %d", 1)
	}
}

// Tests the behavior of ConsumeItem
func TestConsumeItem(t *testing.T) {
	seedDb()
	if err := fakeCommander.ConsumeItem(20001, 5); err != nil {
		t.Errorf("Expected consume item, has %d, need %d", fakeCommander.CommanderItemsMap[20001].Count, 5)
	}
	if err := fakeCommander.ConsumeItem(20001, 0); err != nil {
		t.Errorf("Expected not consume item, has %d, need %d", fakeCommander.CommanderItemsMap[20001].Count, 0)
	}
	if err := fakeCommander.ConsumeItem(20001, 1); err == nil {
		t.Errorf("Expected not consume item, has -, need %d", 1)
	}
	if err := fakeCommander.ConsumeItem(20002, 1); err == nil {
		t.Errorf("Expected not consume item, has -, need %d", 1)
	}
	if err := fakeCommander.ConsumeItem(20001, 400); err == nil {
		t.Errorf("Expected not consume item, has %d, need %d", fakeCommander.CommanderItemsMap[20001].Count, 400)
	}
}

// Tests the behavior of ConsumeResource
func TestConsumeResource(t *testing.T) {
	seedDb()
	if err := fakeCommander.ConsumeResource(2, 5); err != nil {
		t.Errorf("Expected consume resource, has %d, need %d", fakeCommander.OwnedResourcesMap[2].Amount, 5)
	}
	if err := fakeCommander.ConsumeResource(2, 0); err != nil {
		t.Errorf("Expected consume resource, has %d, need %d", fakeCommander.OwnedResourcesMap[2].Amount, 0)
	}
	if err := fakeCommander.ConsumeResource(2, 1); err != nil {
		t.Errorf("Expected consume resource, has %d, need %d", fakeCommander.OwnedResourcesMap[2].Amount, 1)
	}
	if err := fakeCommander.ConsumeResource(3, 1); err == nil {
		t.Errorf("Expected not consume resource, has -, need %d", 1)
	}
	if err := fakeCommander.ConsumeResource(2, 1000); err == nil {
		t.Errorf("Expected not consume resource, has %d, need %d", fakeCommander.OwnedResourcesMap[2].Amount, 1000)
	}
}

func TestCommanderLoadFiltersPunishments(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Punishment{})
	clearTable(t, &Commander{})

	commander := Commander{
		CommanderID: 99,
		AccountID:   99,
		Name:        "Punishment Test",
	}
	if err := CreateCommanderRoot(commander.CommanderID, commander.AccountID, commander.Name, 0, 0); err != nil {
		t.Fatalf("failed to create commander: %v", err)
	}

	past := time.Now().Add(-time.Hour)
	future := time.Now().Add(time.Hour)
	punishments := []Punishment{
		{PunishedID: commander.CommanderID, IsPermanent: false, LiftTimestamp: &past},
		{PunishedID: commander.CommanderID, IsPermanent: false, LiftTimestamp: &future},
	}
	for i := range punishments {
		if err := punishments[i].Create(); err != nil {
			t.Fatalf("failed to create punishments: %v", err)
		}
	}

	loaded := Commander{CommanderID: commander.CommanderID}
	if err := loaded.Load(); err != nil {
		t.Fatalf("failed to load commander: %v", err)
	}
	if len(loaded.Punishments) != 1 {
		t.Fatalf("expected 1 active punishment, got %d", len(loaded.Punishments))
	}
	if loaded.Punishments[0].LiftTimestamp == nil || !loaded.Punishments[0].LiftTimestamp.Equal(future) {
		t.Fatalf("expected active punishment to be the future one")
	}

	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `DELETE FROM commanders WHERE commander_id = $1`, int64(commander.CommanderID)); err != nil {
		t.Fatalf("failed to cleanup commander: %v", err)
	}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `DELETE FROM punishments WHERE punished_id = $1`, int64(commander.CommanderID)); err != nil {
		t.Fatalf("failed to cleanup punishments: %v", err)
	}
}

// Tests the behavior of AddItem
func TestAddItem(t *testing.T) {
	seedDb()
	base := fakeCommander.CommanderItemsMap[20001].Count
	if err := fakeCommander.AddItem(20001, 5); err != nil {
		t.Errorf("Attempt to add %d items (id: %d) failed", 5, 20001)
	}
	if fakeCommander.CommanderItemsMap[20001].Count != base+5 {
		t.Errorf("Count mismatch, has %d, need %d", fakeCommander.CommanderItemsMap[20001].Count, base+5)
	} else {
		base += 5
	}
	if err := fakeCommander.AddItem(20001, 0); err != nil {
		t.Errorf("Attempt to add %d items (id: %d) failed", 0, 20001)
	}
	if fakeCommander.CommanderItemsMap[20001].Count != base {
		t.Errorf("Count mismatch, has %d, need %d", fakeCommander.CommanderItemsMap[20001].Count, base)
	}
	if err := fakeCommander.AddItem(20002, 1); err != nil {
		t.Errorf("Attempt to add %d items (id: %d) failed", 1, 20002)
	}
}

// Tests the behavior of AddResource
func TestAddResource(t *testing.T) {
	seedDb()
	base := fakeCommander.OwnedResourcesMap[2].Amount
	if err := fakeCommander.AddResource(2, 5); err != nil {
		t.Errorf("Attempt to add %d resources (id: %d) failed", 5, 2)
	}
	if fakeCommander.OwnedResourcesMap[2].Amount != base+5 {
		t.Errorf("Count mismatch, has %d, need %d", fakeCommander.OwnedResourcesMap[2].Amount, base+5)
	} else {
		base += 5
	}
	if err := fakeCommander.AddResource(2, 0); err != nil {
		t.Errorf("Attempt to add %d resources (id: %d) failed", 0, 2)
	}
	if fakeCommander.OwnedResourcesMap[2].Amount != base {
		t.Errorf("Count mismatch, has %d, need %d", fakeCommander.OwnedResourcesMap[2].Amount, base)
	}
	if err := fakeCommander.AddResource(3, 1); err != nil {
		t.Errorf("Attempt to add %d resources (id: %d) failed", 1, 3)
	}
}

// Test set resource
func TestSetResource(t *testing.T) {
	seedDb()
	if err := fakeCommander.SetResource(2, 10); err != nil {
		t.Errorf("Attempt to set resource failed")
	}
	if fakeCommander.OwnedResourcesMap[2].Amount != 10 {
		t.Errorf("Count mismatch, has %d, need %d", fakeCommander.OwnedResourcesMap[2].Amount, 10)
	}
	if err := fakeCommander.SetResource(2, 0); err != nil {
		t.Errorf("Attempt to set resource failed")
	}
	if fakeCommander.OwnedResourcesMap[2].Amount != 0 {
		t.Errorf("Count mismatch, has %d, need %d", fakeCommander.OwnedResourcesMap[2].Amount, 0)
	}
	if err := fakeCommander.SetResource(999, 1); err != nil {
		t.Errorf("Attempt to set resource %d failed", 1)
	}
}

// Test set item
func TestSetItem(t *testing.T) {
	seedDb()
	if err := fakeCommander.SetItem(20001, 10); err != nil {
		t.Errorf("Attempt to set item failed")
	}
	if fakeCommander.CommanderItemsMap[20001].Count != 10 {
		t.Errorf("Count mismatch, has %d, need %d", fakeCommander.CommanderItemsMap[20001].Count, 10)
	}
	if err := fakeCommander.SetItem(20001, 0); err != nil {
		t.Errorf("Attempt to set item failed")
	}
	if fakeCommander.CommanderItemsMap[20001].Count != 0 {
		t.Errorf("Count mismatch, has %d, need %d", fakeCommander.CommanderItemsMap[20001].Count, 0)
	}
	if err := fakeCommander.SetItem(20001, 1); err != nil {
		t.Errorf("Attempt to set item %d failed", 20001)
	}
}

// Test the behavior of orm.Commander.GetItemCount
func TestGetItemCount(t *testing.T) {
	seedDb()
	if err := fakeCommander.SetItem(20001, 1); err != nil {
		t.Fatalf("failed to set item count: %v", err)
	}
	if fakeCommander.GetItemCount(20001) != 1 {
		t.Errorf("Count mismatch, has %d, expected %d", fakeCommander.GetItemCount(20001), 1)
	}
	if fakeCommander.GetItemCount(8546213) != 0 {
		t.Errorf("Count mismatch, has %d, expected %d", fakeCommander.GetItemCount(8546213), 0)
	}
}
