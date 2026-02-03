package orm

import (
	"testing"

	"gorm.io/gorm"
)

func TestOwnedEquipmentSetAndRemove(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &OwnedEquipment{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 2001, AccountID: 2001, Name: "Equip Owner"}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	if err := GormDB.Transaction(func(tx *gorm.DB) error {
		return commander.SetOwnedEquipmentTx(tx, 3001, 2)
	}); err != nil {
		t.Fatalf("set owned equipment: %v", err)
	}

	entry := commander.GetOwnedEquipment(3001)
	if entry == nil || entry.Count != 2 {
		t.Fatalf("expected equipment count 2, got %v", entry)
	}

	if err := GormDB.Transaction(func(tx *gorm.DB) error {
		return commander.RemoveOwnedEquipmentTx(tx, 3001, 1)
	}); err != nil {
		t.Fatalf("remove owned equipment: %v", err)
	}
	entry = commander.GetOwnedEquipment(3001)
	if entry == nil || entry.Count != 1 {
		t.Fatalf("expected equipment count 1, got %v", entry)
	}

	if err := GormDB.Transaction(func(tx *gorm.DB) error {
		return commander.SetOwnedEquipmentTx(tx, 3001, 0)
	}); err != nil {
		t.Fatalf("delete owned equipment: %v", err)
	}
	if commander.GetOwnedEquipment(3001) != nil {
		t.Fatalf("expected equipment to be deleted")
	}
}
