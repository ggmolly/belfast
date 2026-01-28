package orm

import (
	"errors"
	"sync"
	"testing"
	"time"

	"gorm.io/gorm"
)

var commanderTestOnce sync.Once

func initCommanderTest(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	commanderTestOnce.Do(func() {
		InitDatabase()
	})
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&Commander{}).Error; err != nil {
		t.Fatalf("clear commanders: %v", err)
	}
}

func TestCommanderCreate(t *testing.T) {
	initCommanderTest(t)

	commander := Commander{
		AccountID:     1001,
		CommanderID:   1001,
		Name:          "Test Commander",
		Level:         10,
		Exp:           5000,
		GuideIndex:    5,
		NewGuideIndex: 5,
		LastLogin:     time.Now().UTC(),
	}

	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander failed: %v", err)
	}

	var stored Commander
	if err := GormDB.Where("commander_id = ?", 1001).First(&stored).Error; err != nil {
		t.Fatalf("fetch commander failed: %v", err)
	}

	if stored.Name != "Test Commander" {
		t.Fatalf("expected name Test Commander, got %s", stored.Name)
	}
	if stored.Level != 10 {
		t.Fatalf("expected level 10, got %d", stored.Level)
	}
}

func TestCommanderRead(t *testing.T) {
	initCommanderTest(t)

	commander := Commander{
		AccountID:   1002,
		CommanderID: 1002,
		Name:        "Read Test",
		Level:       15,
	}

	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander failed: %v", err)
	}

	var read Commander
	if err := GormDB.Where("commander_id = ?", 1002).First(&read).Error; err != nil {
		t.Fatalf("read commander failed: %v", err)
	}

	if read.Name != "Read Test" {
		t.Fatalf("expected name Read Test, got %s", read.Name)
	}
}

func TestCommanderUpdate(t *testing.T) {
	initCommanderTest(t)

	commander := Commander{
		AccountID:   1003,
		CommanderID: 1003,
		Name:        "Update Test",
		Level:       5,
	}

	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander failed: %v", err)
	}

	commander.Level = 20
	commander.Exp = 10000

	if err := GormDB.Save(&commander).Error; err != nil {
		t.Fatalf("update commander failed: %v", err)
	}

	var updated Commander
	if err := GormDB.Where("commander_id = ?", 1003).First(&updated).Error; err != nil {
		t.Fatalf("fetch updated commander failed: %v", err)
	}

	if updated.Level != 20 {
		t.Fatalf("expected level 20, got %d", updated.Level)
	}
	if updated.Exp != 10000 {
		t.Fatalf("expected exp 10000, got %d", updated.Exp)
	}
}

func TestCommanderDelete(t *testing.T) {
	initCommanderTest(t)

	commander := Commander{
		AccountID:   1004,
		CommanderID: 1004,
		Name:        "Delete Test",
		Level:       1,
	}

	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander failed: %v", err)
	}

	if err := GormDB.Delete(&commander).Error; err != nil {
		t.Fatalf("delete commander failed: %v", err)
	}

	var deleted Commander
	err := GormDB.Where("commander_id = ?", 1004).First(&deleted).Error
	if err == nil {
		t.Fatalf("expected commander to be deleted")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestCommanderFindByAccountID(t *testing.T) {
	initCommanderTest(t)

	commander1 := Commander{
		AccountID:   2001,
		CommanderID: 2001,
		Name:        "Commander One",
		Level:       10,
	}

	commander2 := Commander{
		AccountID:   2002,
		CommanderID: 2002,
		Name:        "Commander Two",
		Level:       15,
	}

	GormDB.Create(&commander1)
	GormDB.Create(&commander2)

	var commanders []Commander
	if err := GormDB.Where("account_id = ?", 2001).Find(&commanders).Error; err != nil {
		t.Fatalf("find by account id failed: %v", err)
	}

	if len(commanders) != 1 {
		t.Fatalf("expected 1 commander, got %d", len(commanders))
	}
	if commanders[0].Name != "Commander One" {
		t.Fatalf("expected Commander One, got %s", commanders[0].Name)
	}
}

func TestCommanderNameUniqueness(t *testing.T) {
	initCommanderTest(t)

	commander1 := Commander{
		AccountID:   3001,
		CommanderID: 3001,
		Name:        "Duplicate Name",
		Level:       5,
	}

	GormDB.Create(&commander1)

	commander2 := Commander{
		AccountID:   3002,
		CommanderID: 3002,
		Name:        "Duplicate Name",
		Level:       8,
	}

	err := GormDB.Create(&commander2).Error
	if err == nil {
		t.Fatalf("expected duplicate name to fail")
	}
}

func TestCommanderLevelClamping(t *testing.T) {
	initCommanderTest(t)

	commander := Commander{
		AccountID:   4001,
		CommanderID: 4001,
		Name:        "Level Test",
		Level:       150,
	}

	if err := GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander failed: %v", err)
	}

	var stored Commander
	if err := GormDB.Where("commander_id = ?", 4001).First(&stored).Error; err != nil {
		t.Fatalf("fetch commander failed: %v", err)
	}

	if stored.Level > 120 {
		t.Fatalf("expected level to be clamped to 120, got %d", stored.Level)
	}
}
