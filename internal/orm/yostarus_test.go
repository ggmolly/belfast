package orm

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
)

var yostarusMapTestOnce sync.Once

func initYostarusMapTest(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	yostarusMapTestOnce.Do(func() {
		InitDatabase()
	})
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `DELETE FROM yostarus_maps`); err != nil {
		t.Fatalf("clear yostarus maps: %v", err)
	}
}

func TestYostarusMapCreate(t *testing.T) {
	initYostarusMapTest(t)

	yostarusMap := YostarusMap{
		Arg2:      123456,
		AccountID: 789,
	}

	if err := CreateYostarusMap(yostarusMap.Arg2, yostarusMap.AccountID); err != nil {
		t.Fatalf("create yostarus map failed: %v", err)
	}

	stored, err := GetYostarusMapByArg2(123456)
	if err != nil {
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

	if err := CreateYostarusMap(yostarusMap1.Arg2, yostarusMap1.AccountID); err != nil {
		t.Fatalf("create first yostarus map failed: %v", err)
	}

	err := CreateYostarusMap(yostarusMap2.Arg2, yostarusMap2.AccountID)
	if err == nil {
		t.Fatalf("expected duplicate arg2 to fail")
	}

	stored, err := GetYostarusMapByArg2(111111)
	if err != nil {
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

	if err := CreateYostarusMap(yostarusMap.Arg2, yostarusMap.AccountID); err != nil {
		t.Fatalf("create yostarus map failed: %v", err)
	}

	yostarusMap.AccountID = 777

	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `
UPDATE yostarus_maps
SET account_id = $2
WHERE arg2 = $1
`, int64(yostarusMap.Arg2), int64(yostarusMap.AccountID)); err != nil {
		t.Fatalf("update yostarus map failed: %v", err)
	}

	stored, err := GetYostarusMapByArg2(555555)
	if err != nil {
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

	if err := CreateYostarusMap(yostarusMap.Arg2, yostarusMap.AccountID); err != nil {
		t.Fatalf("create yostarus map failed: %v", err)
	}

	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `DELETE FROM yostarus_maps WHERE arg2 = $1`, int64(yostarusMap.Arg2)); err != nil {
		t.Fatalf("delete yostarus map failed: %v", err)
	}

	_, err := GetYostarusMapByArg2(888888)
	if !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestYostarusMapFind(t *testing.T) {
	initYostarusMapTest(t)

	yostarusMap := YostarusMap{
		Arg2:      999888,
		AccountID: 888,
	}

	if err := CreateYostarusMap(yostarusMap.Arg2, yostarusMap.AccountID); err != nil {
		t.Fatalf("create yostarus map failed: %v", err)
	}

	found, err := GetYostarusMapByArg2(999888)
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

	found, err := GetYostarusMapByArg2(999999)
	if !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}

	if found != nil {
		t.Fatalf("expected nil map for not found, got %+v", found)
	}
}
