package orm

import (
	"encoding/json"
	"sync"
	"testing"

	"gorm.io/gorm"
)

var equipmentTestOnce sync.Once

func initEquipmentTest(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	equipmentTestOnce.Do(func() {
		InitDatabase()
	})
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Equipment{}).Error; err != nil {
		t.Fatalf("clear equipment: %v", err)
	}
}

func TestEquipmentCreate(t *testing.T) {
	initEquipmentTest(t)

	destroyItem := json.RawMessage(`[{"id":1,"count":10}]`)
	restoreItem := json.RawMessage(`[{"id":2,"count":5}]`)

	equipment := Equipment{
		ID:           10001,
		Base:         uint32Ptr(100),
		DestroyGold:  100,
		DestroyItem:  destroyItem,
		EquipLimit:   5,
		Group:        1,
		Important:    1,
		Level:        1,
		Next:         10002,
		Prev:         10000,
		RestoreGold:  50,
		RestoreItem:  restoreItem,
		TransUseGold: 25,
		Type:         1,
	}

	if err := GormDB.Create(&equipment).Error; err != nil {
		t.Fatalf("create equipment: %v", err)
	}

	if equipment.ID != 10001 {
		t.Fatalf("expected id 10001, got %d", equipment.ID)
	}
	if equipment.Level != 1 {
		t.Fatalf("expected level 1, got %d", equipment.Level)
	}
}

func TestEquipmentFind(t *testing.T) {
	initEquipmentTest(t)

	equipment := Equipment{
		ID:           10002,
		DestroyGold:  100,
		EquipLimit:   3,
		Group:        2,
		Important:    1,
		Level:        5,
		Next:         10003,
		Prev:         10001,
		RestoreGold:  50,
		TransUseGold: 25,
		Type:         2,
	}
	GormDB.Create(&equipment)

	var found Equipment
	if err := GormDB.First(&found, equipment.ID).Error; err != nil {
		t.Fatalf("find equipment: %v", err)
	}

	if found.ID != 10002 {
		t.Fatalf("expected id 10002, got %d", found.ID)
	}
	if found.Level != 5 {
		t.Fatalf("expected level 5, got %d", found.Level)
	}
}

func TestEquipmentUpdate(t *testing.T) {
	initEquipmentTest(t)

	equipment := Equipment{
		ID:           10003,
		DestroyGold:  100,
		EquipLimit:   5,
		Group:        1,
		Important:    1,
		Level:        1,
		RestoreGold:  50,
		TransUseGold: 25,
		Type:         1,
	}
	GormDB.Create(&equipment)

	equipment.Level = 10
	equipment.EquipLimit = 8

	if err := GormDB.Save(&equipment).Error; err != nil {
		t.Fatalf("update equipment: %v", err)
	}

	var found Equipment
	if err := GormDB.First(&found, equipment.ID).Error; err != nil {
		t.Fatalf("find updated equipment: %v", err)
	}

	if found.Level != 10 {
		t.Fatalf("expected level 10, got %d", found.Level)
	}
	if found.EquipLimit != 8 {
		t.Fatalf("expected equip limit 8, got %d", found.EquipLimit)
	}
}

func TestEquipmentDelete(t *testing.T) {
	initEquipmentTest(t)

	equipment := Equipment{
		ID:           10004,
		DestroyGold:  100,
		EquipLimit:   5,
		Group:        1,
		Important:    1,
		Level:        1,
		RestoreGold:  50,
		TransUseGold: 25,
		Type:         1,
	}
	GormDB.Create(&equipment)

	if err := GormDB.Delete(&equipment).Error; err != nil {
		t.Fatalf("delete equipment: %v", err)
	}

	var found Equipment
	err := GormDB.First(&found, equipment.ID).Error
	if err != gorm.ErrRecordNotFound {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestEquipmentOptionalBase(t *testing.T) {
	initEquipmentTest(t)

	t.Run("base nil", func(t *testing.T) {
		equipment := Equipment{
			ID:           20001,
			DestroyGold:  100,
			EquipLimit:   5,
			Group:        1,
			Important:    1,
			Level:        1,
			RestoreGold:  50,
			TransUseGold: 25,
			Type:         1,
		}
		GormDB.Create(&equipment)

		var found Equipment
		if err := GormDB.First(&found, equipment.ID).Error; err != nil {
			t.Fatalf("find equipment: %v", err)
		}

		if found.Base != nil {
			t.Fatalf("expected Base to be nil, got %v", found.Base)
		}
	})

	t.Run("base set", func(t *testing.T) {
		base := uint32(200)
		equipment := Equipment{
			ID:           20002,
			Base:         &base,
			DestroyGold:  100,
			EquipLimit:   5,
			Group:        1,
			Important:    1,
			Level:        1,
			RestoreGold:  50,
			TransUseGold: 25,
			Type:         1,
		}
		GormDB.Create(&equipment)

		var found Equipment
		if err := GormDB.First(&found, equipment.ID).Error; err != nil {
			t.Fatalf("find equipment: %v", err)
		}

		if found.Base == nil || *found.Base != 200 {
			t.Fatalf("expected Base to be 200, got %v", found.Base)
		}
	})
}

func TestEquipmentNextPrev(t *testing.T) {
	initEquipmentTest(t)

	equipment := Equipment{
		ID:           30001,
		DestroyGold:  100,
		EquipLimit:   5,
		Group:        1,
		Important:    1,
		Level:        1,
		Next:         30002,
		Prev:         30000,
		RestoreGold:  50,
		TransUseGold: 25,
		Type:         1,
	}
	GormDB.Create(&equipment)

	var found Equipment
	if err := GormDB.First(&found, equipment.ID).Error; err != nil {
		t.Fatalf("find equipment: %v", err)
	}

	if found.Next != 30002 {
		t.Fatalf("expected next 30002, got %d", found.Next)
	}
	if found.Prev != 30000 {
		t.Fatalf("expected prev 30000, got %d", found.Prev)
	}
}

func TestEquipmentJSONFields(t *testing.T) {
	initEquipmentTest(t)

	destroyItem := json.RawMessage(`[{"id":100,"count":5},{"id":101,"count":3}]`)
	restoreItem := json.RawMessage(`[{"id":200,"count":2}]`)

	equipment := Equipment{
		ID:           40001,
		DestroyGold:  100,
		DestroyItem:  destroyItem,
		EquipLimit:   5,
		Group:        1,
		Important:    1,
		Level:        1,
		RestoreGold:  50,
		RestoreItem:  restoreItem,
		TransUseGold: 25,
		Type:         1,
	}
	GormDB.Create(&equipment)

	var found Equipment
	if err := GormDB.First(&found, equipment.ID).Error; err != nil {
		t.Fatalf("find equipment: %v", err)
	}

	if found.DestroyItem == nil {
		t.Fatalf("expected DestroyItem to be set, got nil")
	}
	if found.RestoreItem == nil {
		t.Fatalf("expected RestoreItem to be set, got nil")
	}
}

func TestEquipmentGroup(t *testing.T) {
	initEquipmentTest(t)

	equipment := Equipment{
		ID:           50001,
		DestroyGold:  100,
		EquipLimit:   5,
		Group:        99,
		Important:    1,
		Level:        1,
		RestoreGold:  50,
		TransUseGold: 25,
		Type:         1,
	}
	GormDB.Create(&equipment)

	var found Equipment
	if err := GormDB.First(&found, equipment.ID).Error; err != nil {
		t.Fatalf("find equipment: %v", err)
	}

	if found.Group != 99 {
		t.Fatalf("expected group 99, got %d", found.Group)
	}
}

func TestEquipmentImportant(t *testing.T) {
	initEquipmentTest(t)

	tests := []struct {
		name      string
		important uint32
	}{
		{"not important", 0},
		{"important", 1},
		{"very important", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			equipment := Equipment{
				ID:           uint32(60001 + tt.important),
				DestroyGold:  100,
				EquipLimit:   5,
				Group:        1,
				Important:    tt.important,
				Level:        1,
				RestoreGold:  50,
				TransUseGold: 25,
				Type:         1,
			}
			GormDB.Create(&equipment)

			var found Equipment
			if err := GormDB.First(&found, equipment.ID).Error; err != nil {
				t.Fatalf("find equipment: %v", err)
			}

			if found.Important != tt.important {
				t.Fatalf("expected important %d, got %d", tt.important, found.Important)
			}
		})
	}
}

func uint32Ptr(v uint32) *uint32 {
	return &v
}
