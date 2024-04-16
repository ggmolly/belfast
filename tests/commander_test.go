package tests

import (
	"os"
	"testing"

	"github.com/ggmolly/belfast/orm"
)

var fakeCommander orm.Commander

var (
	fakeResources []orm.Resource
	fakeItems     []orm.Item
)

func seedDb() {
	tx := orm.GormDB.Begin()
	for _, r := range fakeResources {
		tx.Save(&r)
	}
	for _, i := range fakeItems {
		tx.Save(&i)
	}
	if err := tx.Commit().Error; err != nil {
		panic(err)
	}
}

func init() {
	fakeCommander.AccountID = 1
	fakeCommander.CommanderID = 1
	fakeCommander.Name = "Fake Commander"
	fakeResources = []orm.Resource{
		{ID: 1, Name: "Gold"},
		{ID: 2, Name: "Fake resource"},
	}
	fakeItems = []orm.Item{
		{ID: 20001, Name: "Wisdom Cube"},
		{ID: 45, Name: "Fake Item"},
		{ID: 60, Name: "Fake Item 2"},
	}

	// Init the database
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	seedDb()

	fakeCommander.OwnedResourcesMap = make(map[uint32]*orm.OwnedResource)
	fakeCommander.CommanderItemsMap = make(map[uint32]*orm.CommanderItem)

	tx := orm.GormDB.Begin()
	// Fake resources
	fakeResourcesCnt := []uint32{100, 30}
	for i := 0; i < len(fakeResources); i++ {
		resource := orm.OwnedResource{
			ResourceID:  fakeResources[i].ID,
			Amount:      fakeResourcesCnt[i],
			CommanderID: fakeCommander.CommanderID,
		}
		tx.Create(&resource)
		fakeCommander.OwnedResources = append(fakeCommander.OwnedResources, resource)
		fakeCommander.OwnedResourcesMap[fakeResources[i].ID] = &fakeCommander.OwnedResources[i]
	}

	// Fake items
	fakeItemsCnt := []uint32{5, 50, 3}

	for i := 0; i < len(fakeItems); i++ {
		item := orm.CommanderItem{
			ItemID:      fakeItems[i].ID,
			Count:       fakeItemsCnt[i],
			CommanderID: fakeCommander.CommanderID,
		}
		tx.Create(&item)
		fakeCommander.Items = append(fakeCommander.Items, item)
		fakeCommander.CommanderItemsMap[fakeItems[i].ID] = &fakeCommander.Items[i]
	}

	tx.Create(&fakeCommander)
	if err := tx.Commit().Error; err != nil {
		panic(err)
	}
}

// Tests the behavior of orm.Commander.HasEnoughGold
func TestEnoughGold(t *testing.T) {
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
	if fakeCommander.GetItemCount(20001) != 1 {
		t.Errorf("Count mismatch, has %d, expected %d", fakeCommander.GetItemCount(20001), 1)
	}
	if fakeCommander.GetItemCount(8546213) != 0 {
		t.Errorf("Count mismatch, has %d, expected %d", fakeCommander.GetItemCount(8546213), 0)
	}
}
