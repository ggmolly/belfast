package orm

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
)

var deviceAuthMapTestOnce sync.Once

func initDeviceAuthMapTest(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	deviceAuthMapTestOnce.Do(func() {
		InitDatabase()
	})
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `DELETE FROM device_auth_maps`); err != nil {
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

	if err := UpsertDeviceAuthMap(authMap.DeviceID, authMap.Arg2, authMap.AccountID); err != nil {
		t.Fatalf("create device auth map failed: %v", err)
	}

	stored, err := GetDeviceAuthMapByDeviceID("device-1001")
	if err != nil {
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

	if err := UpsertDeviceAuthMap(authMap.DeviceID, authMap.Arg2, authMap.AccountID); err != nil {
		t.Fatalf("create device auth map failed: %v", err)
	}

	stored, err := GetDeviceAuthMapByDeviceID("device-1002")
	if err != nil {
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

	if err := UpsertDeviceAuthMap(authMap.DeviceID, authMap.Arg2, authMap.AccountID); err != nil {
		t.Fatalf("create device auth map failed: %v", err)
	}

	var stored DeviceAuthMap
	if err := db.DefaultStore.Pool.QueryRow(context.Background(), `
SELECT device_id, arg2, account_id
FROM device_auth_maps
WHERE arg2 = $1
`, int64(345678)).Scan(&stored.DeviceID, &stored.Arg2, &stored.AccountID); err != nil {
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

	if err := UpsertDeviceAuthMap(authMap.DeviceID, authMap.Arg2, authMap.AccountID); err != nil {
		t.Fatalf("create device auth map failed: %v", err)
	}

	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `DELETE FROM device_auth_maps WHERE device_id = $1`, authMap.DeviceID); err != nil {
		t.Fatalf("delete device auth map failed: %v", err)
	}

	_, err := GetDeviceAuthMapByDeviceID("device-1004")
	if !errors.Is(err, db.ErrNotFound) {
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

	if err := UpsertDeviceAuthMap(authMap.DeviceID, authMap.Arg2, authMap.AccountID); err != nil {
		t.Fatalf("create device auth map failed: %v", err)
	}

	authMap.Arg2 = 999999

	if err := UpsertDeviceAuthMap(authMap.DeviceID, authMap.Arg2, authMap.AccountID); err != nil {
		t.Fatalf("update device auth map failed: %v", err)
	}

	stored, err := GetDeviceAuthMapByDeviceID("device-1005")
	if err != nil {
		t.Fatalf("fetch updated device auth map failed: %v", err)
	}

	if stored.Arg2 != 999999 {
		t.Fatalf("expected arg2 999999, got %d", stored.Arg2)
	}
}
