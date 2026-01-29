package orm

import (
	"errors"
	"testing"

	"github.com/ggmolly/belfast/internal/rng"
	"gorm.io/gorm"
)

func TestShipCRUDAndValidate(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Ship{})
	clearTable(t, &Rarity{})

	if err := GormDB.Create(&Rarity{ID: 2, Name: "Common"}).Error; err != nil {
		t.Fatalf("seed rarity: %v", err)
	}
	ship := Ship{TemplateID: 5001, Name: "Ship", EnglishName: "Ship", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	if err := ship.Create(); err != nil {
		t.Fatalf("create ship: %v", err)
	}
	ship.Name = "Updated"
	if err := ship.Update(); err != nil {
		t.Fatalf("update ship: %v", err)
	}
	loaded := Ship{TemplateID: ship.TemplateID}
	if err := loaded.Retrieve(false); err != nil {
		t.Fatalf("retrieve ship: %v", err)
	}
	if err := loaded.Retrieve(true); err != nil {
		t.Fatalf("retrieve ship greedy: %v", err)
	}
	if err := ValidateShipID(ship.TemplateID); err != nil {
		t.Fatalf("validate ship id: %v", err)
	}
	if err := ValidateShipID(9999); err == nil {
		t.Fatalf("expected ship not found error")
	}
	if err := loaded.Delete(); err != nil {
		t.Fatalf("delete ship: %v", err)
	}
	if err := loaded.Retrieve(false); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestGetRandomPoolShip(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Ship{})

	originalRng := shipRng
	shipRng = rng.NewLockedRandFromSeed(2)
	defer func() { shipRng = originalRng }()

	poolID := uint32(1)
	for _, rarity := range []uint32{2, 3, 4, 5} {
		ship := Ship{TemplateID: rarity + 6000, Name: "Ship", EnglishName: "Ship", RarityID: rarity, Star: 1, Type: 1, Nationality: 1, BuildTime: 10, PoolID: &poolID}
		if err := GormDB.Create(&ship).Error; err != nil {
			t.Fatalf("seed pool ship: %v", err)
		}
	}
	selected, err := GetRandomPoolShip(poolID)
	if err != nil {
		t.Fatalf("get random pool ship: %v", err)
	}
	if selected.PoolID == nil || *selected.PoolID != poolID {
		t.Fatalf("expected pool id %d", poolID)
	}
}
