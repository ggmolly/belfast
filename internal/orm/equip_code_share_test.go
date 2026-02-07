package orm

import (
	"sync"
	"testing"

	"gorm.io/gorm"
)

var equipCodeShareTestOnce sync.Once

func initEquipCodeShareTest(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	equipCodeShareTestOnce.Do(func() {
		InitDatabase()
	})
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&EquipCodeShare{}).Error; err != nil {
		t.Fatalf("clear equip code shares: %v", err)
	}
}

func TestEquipCodeShareCreate(t *testing.T) {
	initEquipCodeShareTest(t)
	share := EquipCodeShare{CommanderID: 1, ShipGroupID: 2, ShareDay: 10}
	if err := GormDB.Create(&share).Error; err != nil {
		t.Fatalf("create share failed: %v", err)
	}
	var stored EquipCodeShare
	if err := GormDB.Where("commander_id = ? AND ship_group_id = ? AND share_day = ?", 1, 2, 10).First(&stored).Error; err != nil {
		t.Fatalf("fetch share failed: %v", err)
	}
	if stored.CommanderID != 1 {
		t.Fatalf("expected commander_id 1, got %d", stored.CommanderID)
	}
}

func TestEquipCodeShareDedupeIndex(t *testing.T) {
	initEquipCodeShareTest(t)
	first := EquipCodeShare{CommanderID: 2, ShipGroupID: 3, ShareDay: 11}
	second := EquipCodeShare{CommanderID: 2, ShipGroupID: 3, ShareDay: 11}
	if err := GormDB.Create(&first).Error; err != nil {
		t.Fatalf("create share failed: %v", err)
	}
	if err := GormDB.Create(&second).Error; err == nil {
		t.Fatalf("expected duplicate share insert to fail")
	}
}
