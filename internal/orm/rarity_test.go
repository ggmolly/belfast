package orm

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
)

var rarityTestOnce sync.Once

func initRarityTest(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	rarityTestOnce.Do(func() {
		InitDatabase()
	})
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `DELETE FROM rarities`); err != nil {
		t.Fatalf("clear rarities: %v", err)
	}
}

func TestRarityCreate(t *testing.T) {
	initRarityTest(t)

	rarity := Rarity{
		ID:   2,
		Name: "Common",
	}

	if err := CreateRarity(&rarity); err != nil {
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
		if err := CreateRarity(&rarity); err != nil {
			t.Fatalf("create rarity %d: %v", rarity.ID, err)
		}
	}

	found, _, err := ListRarities(0, 100)
	if err != nil {
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
	if err := CreateRarity(&rarity); err != nil {
		t.Fatalf("create rarity: %v", err)
	}

	found, err := GetRarityByID(rarity.ID)
	if err != nil {
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
	if err := CreateRarity(&rarity); err != nil {
		t.Fatalf("create rarity: %v", err)
	}

	rarity.Name = "Updated Rare"
	if err := UpdateRarity(&rarity); err != nil {
		t.Fatalf("update rarity: %v", err)
	}

	found, err := GetRarityByID(rarity.ID)
	if err != nil {
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
	if err := CreateRarity(&rarity); err != nil {
		t.Fatalf("create rarity: %v", err)
	}

	if err := DeleteRarity(rarity.ID); err != nil {
		t.Fatalf("delete rarity: %v", err)
	}

	_, err := GetRarityByID(rarity.ID)
	if !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestRarityNameLength(t *testing.T) {
	initRarityTest(t)

	rarity := Rarity{
		ID:   1,
		Name: "SuperDuperUltraMegaRare",
	}
	if err := CreateRarity(&rarity); err != nil {
		t.Fatalf("create rarity: %v", err)
	}

	found, err := GetRarityByID(rarity.ID)
	if err != nil {
		t.Fatalf("find rarity: %v", err)
	}

	if len(found.Name) != len(rarity.Name) {
		t.Fatalf("expected name length %d, got %d", len(rarity.Name), len(found.Name))
	}
}
