package orm

import (
	"errors"
	"sync"
	"testing"

	"gorm.io/gorm"
)

var deviceAuthMapTestOnce sync.Once

func initDeviceAuthMapTest(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	deviceAuthMapTestOnce.Do(func() {
		InitDatabase()
	})
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&DeviceAuthMap{}).Error; err != nil {
		t.Fatalf("clear device auth maps: %v", err)
	}
}

func TestDeviceAuthMapCreate(t *testing.T) {
	initDeviceAuthMapTest(t)

	authMap := DeviceAuthMap{
		DeviceID:  "device-1001",
		Arg2:      123456,
		AccountID: 789,
	}

	if err := GormDB.Create(&authMap).Error; err != nil {
		t.Fatalf("create device auth map failed: %v", err)
	}

	var stored DeviceAuthMap
	if err := GormDB.Where("device_id = ?", "device-1001").First(&stored).Error; err != nil {
		t.Fatalf("fetch device auth map failed: %v", err)
	}

	if stored.DeviceID != "device-1001" {
		t.Fatalf("expected device id device-1001, got %s", stored.DeviceID)
	}
	if stored.Arg2 != 123456 {
		t.Fatalf("expected arg2 123456, got %d", stored.Arg2)
	}
	if stored.AccountID != 789 {
		t.Fatalf("expected account id 789, got %d", stored.AccountID)
	}
}

func TestDeviceAuthMapReadByDeviceID(t *testing.T) {
	initDeviceAuthMapTest(t)

	authMap := DeviceAuthMap{
		DeviceID:  "device-1002",
		Arg2:      234567,
		AccountID: 890,
	}

	GormDB.Create(&authMap)

	var stored DeviceAuthMap
	if err := GormDB.Where("device_id = ?", "device-1002").First(&stored).Error; err != nil {
		t.Fatalf("fetch by device id failed: %v", err)
	}

	if stored.Arg2 != 234567 {
		t.Fatalf("expected arg2 234567, got %d", stored.Arg2)
	}
}

func TestDeviceAuthMapReadByArg2(t *testing.T) {
	initDeviceAuthMapTest(t)

	authMap := DeviceAuthMap{
		DeviceID:  "device-1003",
		Arg2:      345678,
		AccountID: 901,
	}

	GormDB.Create(&authMap)

	var stored DeviceAuthMap
	if err := GormDB.Where("arg2 = ?", 345678).First(&stored).Error; err != nil {
		t.Fatalf("fetch by arg2 failed: %v", err)
	}

	if stored.DeviceID != "device-1003" {
		t.Fatalf("expected device id device-1003, got %s", stored.DeviceID)
	}
}

func TestDeviceAuthMapDelete(t *testing.T) {
	initDeviceAuthMapTest(t)

	authMap := DeviceAuthMap{
		DeviceID:  "device-1004",
		Arg2:      456789,
		AccountID: 912,
	}

	GormDB.Create(&authMap)

	if err := GormDB.Delete(&authMap).Error; err != nil {
		t.Fatalf("delete device auth map failed: %v", err)
	}

	var stored DeviceAuthMap
	err := GormDB.Where("device_id = ?", "device-1004").First(&stored).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestDeviceAuthMapUpdate(t *testing.T) {
	initDeviceAuthMapTest(t)

	authMap := DeviceAuthMap{
		DeviceID:  "device-1005",
		Arg2:      567890,
		AccountID: 923,
	}

	GormDB.Create(&authMap)

	authMap.Arg2 = 999999

	if err := GormDB.Save(&authMap).Error; err != nil {
		t.Fatalf("update device auth map failed: %v", err)
	}

	var stored DeviceAuthMap
	if err := GormDB.Where("device_id = ?", "device-1005").First(&stored).Error; err != nil {
		t.Fatalf("fetch updated device auth map failed: %v", err)
	}

	if stored.Arg2 != 999999 {
		t.Fatalf("expected arg2 999999, got %d", stored.Arg2)
	}
}
