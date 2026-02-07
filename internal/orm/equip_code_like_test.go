package orm

import (
	"sync"
	"testing"

	"gorm.io/gorm"
)

var equipCodeLikeTestOnce sync.Once

func initEquipCodeLikeTest(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	equipCodeLikeTestOnce.Do(func() {
		InitDatabase()
	})
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&EquipCodeLike{}).Error; err != nil {
		t.Fatalf("clear equip code likes: %v", err)
	}
}

func TestEquipCodeLikeCreate(t *testing.T) {
	initEquipCodeLikeTest(t)
	like := EquipCodeLike{CommanderID: 1, ShipGroupID: 2, ShareID: 3, LikeDay: 10}
	if err := GormDB.Create(&like).Error; err != nil {
		t.Fatalf("create like failed: %v", err)
	}

	var stored EquipCodeLike
	if err := GormDB.Where("commander_id = ? AND share_id = ?", 1, 3).First(&stored).Error; err != nil {
		t.Fatalf("fetch like failed: %v", err)
	}
	if stored.ShipGroupID != 2 {
		t.Fatalf("expected shipgroup 2, got %d", stored.ShipGroupID)
	}
}

func TestEquipCodeLikeDedupeIndex(t *testing.T) {
	initEquipCodeLikeTest(t)
	first := EquipCodeLike{CommanderID: 2, ShipGroupID: 3, ShareID: 4, LikeDay: 11}
	second := EquipCodeLike{CommanderID: 2, ShipGroupID: 3, ShareID: 4, LikeDay: 11}
	if err := GormDB.Create(&first).Error; err != nil {
		t.Fatalf("create like failed: %v", err)
	}
	if err := GormDB.Create(&second).Error; err == nil {
		t.Fatalf("expected duplicate like insert to fail")
	}
}

func TestEquipCodeLikeDedupeAllowsDifferentShipGroup(t *testing.T) {
	initEquipCodeLikeTest(t)
	first := EquipCodeLike{CommanderID: 2, ShipGroupID: 3, ShareID: 4, LikeDay: 11}
	second := EquipCodeLike{CommanderID: 2, ShipGroupID: 999, ShareID: 4, LikeDay: 11}
	if err := GormDB.Create(&first).Error; err != nil {
		t.Fatalf("create like failed: %v", err)
	}
	if err := GormDB.Create(&second).Error; err != nil {
		t.Fatalf("create like with different shipgroup failed: %v", err)
	}
}
