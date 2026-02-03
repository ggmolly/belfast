package orm

import (
	"os"
	"testing"

	"gorm.io/gorm"
)

func TestConsumeResourceTx(t *testing.T) {
	os.Setenv("MODE", "test")
	InitDatabase()
	clearTransformTable(t, &OwnedResource{})
	clearTransformTable(t, &Commander{})
	commander := Commander{CommanderID: 700, AccountID: 700, Name: "Resource Tester"}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}
	if err := GormDB.Create(&OwnedResource{CommanderID: commander.CommanderID, ResourceID: 1, Amount: 10}).Error; err != nil {
		t.Fatalf("seed resource: %v", err)
	}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	tx := GormDB.Begin()
	if err := commander.ConsumeResourceTx(tx, 1, 4); err != nil {
		tx.Rollback()
		t.Fatalf("consume resource: %v", err)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}
	var resource OwnedResource
	if err := GormDB.Where("commander_id = ? AND resource_id = ?", commander.CommanderID, 1).First(&resource).Error; err != nil {
		t.Fatalf("load resource: %v", err)
	}
	if resource.Amount != 6 {
		t.Fatalf("expected resource amount 6, got %d", resource.Amount)
	}
}

func TestToProtoOwnedShipTransformList(t *testing.T) {
	os.Setenv("MODE", "test")
	InitDatabase()
	clearTransformTable(t, &OwnedShipTransform{})
	clearTransformTable(t, &OwnedShip{})
	clearTransformTable(t, &Commander{})
	clearTransformTable(t, &Ship{})
	commander := Commander{CommanderID: 701, AccountID: 701, Name: "Transform Tester"}
	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}
	ship := Ship{TemplateID: 9001, Name: "Ship", EnglishName: "Ship", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	if err := GormDB.Create(&ship).Error; err != nil {
		t.Fatalf("create ship: %v", err)
	}
	owned := OwnedShip{OwnerID: commander.CommanderID, ShipID: ship.TemplateID, Level: 1}
	if err := GormDB.Create(&owned).Error; err != nil {
		t.Fatalf("create owned ship: %v", err)
	}
	transform := OwnedShipTransform{OwnerID: commander.CommanderID, ShipID: owned.ID, TransformID: 12011, Level: 1}
	if err := GormDB.Create(&transform).Error; err != nil {
		t.Fatalf("create transform: %v", err)
	}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	loaded := commander.OwnedShipsMap[owned.ID]
	info := ToProtoOwnedShip(*loaded, nil)
	if len(info.TransformList) != 1 {
		t.Fatalf("expected transform list length 1, got %d", len(info.TransformList))
	}
	if info.TransformList[0].GetId() != 12011 || info.TransformList[0].GetLevel() != 1 {
		t.Fatalf("unexpected transform list entry")
	}
}

func clearTransformTable(t *testing.T, model any) {
	t.Helper()
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(model).Error; err != nil {
		t.Fatalf("failed to clear table: %v", err)
	}
}
