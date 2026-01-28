package orm

import (
	"errors"
	"sync"
	"testing"

	"gorm.io/gorm"
)

var yostarusMapTestOnce sync.Once

func initYostarusMapTest(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	yostarusMapTestOnce.Do(func() {
		InitDatabase()
	})
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&YostarusMap{}).Error; err != nil {
		t.Fatalf("clear yostarus maps: %v", err)
	}
}

func TestYostarusMapCreate(t *testing.T) {
	initYostarusMapTest(t)

	yostarusMap := YostarusMap{
		Arg2:      123456,
		AccountID: 789,
	}

	if err := GormDB.Create(&yostarusMap).Error; err != nil {
		t.Fatalf("create yostarus map failed: %v", err)
	}

	var stored YostarusMap
	if err := GormDB.Where("arg2 = ?", 123456).First(&stored).Error; err != nil {
		t.Fatalf("fetch yostarus map failed: %v", err)
	}

	if stored.Arg2 != 123456 {
		t.Fatalf("expected arg2 123456, got %d", stored.Arg2)
	}
	if stored.AccountID != 789 {
		t.Fatalf("expected account id 789, got %d", stored.AccountID)
	}
}

func TestYostarusMapDuplicateArg2(t *testing.T) {
	initYostarusMapTest(t)

	yostarusMap1 := YostarusMap{
		Arg2:      111111,
		AccountID: 111,
	}

	yostarusMap2 := YostarusMap{
		Arg2:      111111,
		AccountID: 222,
	}

	GormDB.Create(&yostarusMap1)

	err := GormDB.Create(&yostarusMap2).Error
	if err == nil {
		t.Fatalf("expected duplicate arg2 to fail")
	}

	var stored YostarusMap
	if err := GormDB.Where("arg2 = ?", 111111).First(&stored).Error; err != nil {
		t.Fatalf("fetch yostarus map failed: %v", err)
	}

	if stored.AccountID != 111 {
		t.Fatalf("expected account id 111 (first), got %d", stored.AccountID)
	}
}

func TestYostarusMapUpdate(t *testing.T) {
	initYostarusMapTest(t)

	yostarusMap := YostarusMap{
		Arg2:      555555,
		AccountID: 666,
	}

	GormDB.Create(&yostarusMap)

	yostarusMap.AccountID = 777

	if err := GormDB.Save(&yostarusMap).Error; err != nil {
		t.Fatalf("update yostarus map failed: %v", err)
	}

	var stored YostarusMap
	if err := GormDB.Where("arg2 = ?", 555555).First(&stored).Error; err != nil {
		t.Fatalf("fetch updated yostarus map failed: %v", err)
	}

	if stored.AccountID != 777 {
		t.Fatalf("expected account id 777, got %d", stored.AccountID)
	}
}

func TestYostarusMapDelete(t *testing.T) {
	initYostarusMapTest(t)

	yostarusMap := YostarusMap{
		Arg2:      888888,
		AccountID: 999,
	}

	GormDB.Create(&yostarusMap)

	if err := GormDB.Delete(&yostarusMap).Error; err != nil {
		t.Fatalf("delete yostarus map failed: %v", err)
	}

	var stored YostarusMap
	err := GormDB.Where("arg2 = ?", 888888).First(&stored).Error

	if err == nil {
		t.Fatalf("expected yostarus map to be deleted")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestYostarusMapFind(t *testing.T) {
	initYostarusMapTest(t)

	yostarusMap := YostarusMap{
		Arg2:      999888,
		AccountID: 888,
	}

	GormDB.Create(&yostarusMap)

	var found YostarusMap
	err := GormDB.Where("arg2 = ?", 999888).First(&found).Error

	if err != nil {
		t.Fatalf("find yostarus map failed: %v", err)
	}

	if found.Arg2 != 999888 {
		t.Fatalf("expected arg2 999888, got %d", found.Arg2)
	}
	if found.AccountID != 888 {
		t.Fatalf("expected account id 888, got %d", found.AccountID)
	}
}

func TestYostarusMapFindNotFound(t *testing.T) {
	initYostarusMapTest(t)

	var found YostarusMap
	err := GormDB.Where("arg2 = ?", 999999).First(&found).Error

	if err != gorm.ErrRecordNotFound {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}

	if found.Arg2 != 0 {
		t.Fatalf("expected arg2 0 for not found, got %d", found.Arg2)
	}
}
