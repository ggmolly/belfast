package orm

import (
	"sync"
	"testing"

	"gorm.io/gorm"
)

var rarityTestOnce sync.Once

func initRarityTest(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	rarityTestOnce.Do(func() {
		InitDatabase()
	})
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Rarity{}).Error; err != nil {
		t.Fatalf("clear rarities: %v", err)
	}
}

func TestRarityCreate(t *testing.T) {
	initRarityTest(t)

	rarity := Rarity{
		ID:   2,
		Name: "Common",
	}

	if err := GormDB.Create(&rarity).Error; err != nil {
		t.Fatalf("create rarity: %v", err)
	}

	if rarity.ID != 2 {
		t.Fatalf("expected id 2, got %d", rarity.ID)
	}
	if rarity.Name != "Common" {
		t.Fatalf("expected name 'Common', got %s", rarity.Name)
	}
}

func TestRarityMultipleRarities(t *testing.T) {
	initRarityTest(t)

	rarities := []Rarity{
		{ID: 2, Name: "Common"},
		{ID: 3, Name: "Rare"},
		{ID: 4, Name: "Elite"},
		{ID: 5, Name: "Super Rare"},
		{ID: 6, Name: "Ultra Rare"},
	}

	for _, rarity := range rarities {
		if err := GormDB.Create(&rarity).Error; err != nil {
			t.Fatalf("create rarity %d: %v", rarity.ID, err)
		}
	}

	var found []Rarity
	if err := GormDB.Find(&found).Error; err != nil {
		t.Fatalf("find rarities: %v", err)
	}

	if len(found) != 5 {
		t.Fatalf("expected 5 rarities, got %d", len(found))
	}
}

func TestRarityFind(t *testing.T) {
	initRarityTest(t)

	rarity := Rarity{
		ID:   5,
		Name: "Super Rare",
	}
	GormDB.Create(&rarity)

	var found Rarity
	if err := GormDB.First(&found, rarity.ID).Error; err != nil {
		t.Fatalf("find rarity: %v", err)
	}

	if found.ID != 5 {
		t.Fatalf("expected id 5, got %d", found.ID)
	}
	if found.Name != "Super Rare" {
		t.Fatalf("expected name 'Super Rare', got %s", found.Name)
	}
}

func TestRarityUpdate(t *testing.T) {
	initRarityTest(t)

	rarity := Rarity{
		ID:   3,
		Name: "Rare",
	}
	GormDB.Create(&rarity)

	rarity.Name = "Updated Rare"
	if err := GormDB.Save(&rarity).Error; err != nil {
		t.Fatalf("update rarity: %v", err)
	}

	var found Rarity
	if err := GormDB.First(&found, rarity.ID).Error; err != nil {
		t.Fatalf("find updated rarity: %v", err)
	}

	if found.Name != "Updated Rare" {
		t.Fatalf("expected name 'Updated Rare', got %s", found.Name)
	}
}

func TestRarityDelete(t *testing.T) {
	initRarityTest(t)

	rarity := Rarity{
		ID:   4,
		Name: "Elite",
	}
	GormDB.Create(&rarity)

	if err := GormDB.Delete(&rarity).Error; err != nil {
		t.Fatalf("delete rarity: %v", err)
	}

	var found Rarity
	err := GormDB.First(&found, rarity.ID).Error
	if err != gorm.ErrRecordNotFound {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestRarityNameLength(t *testing.T) {
	initRarityTest(t)

	rarity := Rarity{
		ID:   1,
		Name: "SuperDuperUltraMegaRare",
	}
	GormDB.Create(&rarity)

	var found Rarity
	if err := GormDB.First(&found, rarity.ID).Error; err != nil {
		t.Fatalf("find rarity: %v", err)
	}

	if len(found.Name) != len(rarity.Name) {
		t.Fatalf("expected name length %d, got %d", len(rarity.Name), len(found.Name))
	}
}
