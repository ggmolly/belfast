package orm

import (
	"sync"
	"testing"

	"gorm.io/gorm"
)

var equipCodeReportTestOnce sync.Once

func initEquipCodeReportTest(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	equipCodeReportTestOnce.Do(func() {
		InitDatabase()
	})
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&EquipCodeReport{}).Error; err != nil {
		t.Fatalf("clear equip code reports: %v", err)
	}
}

func TestEquipCodeReportCreate(t *testing.T) {
	initEquipCodeReportTest(t)
	report := EquipCodeReport{CommanderID: 1, ShipGroupID: 2, ShareID: 3, ReportType: 1, ReportDay: 10}
	if err := GormDB.Create(&report).Error; err != nil {
		t.Fatalf("create report failed: %v", err)
	}
	var stored EquipCodeReport
	if err := GormDB.Where("commander_id = ? AND share_id = ?", 1, 3).First(&stored).Error; err != nil {
		t.Fatalf("fetch report failed: %v", err)
	}
	if stored.ReportType != 1 {
		t.Fatalf("expected report_type 1, got %d", stored.ReportType)
	}
	if stored.ShipGroupID != 2 {
		t.Fatalf("expected shipgroup 2, got %d", stored.ShipGroupID)
	}
}

func TestEquipCodeReportDedupeIndex(t *testing.T) {
	initEquipCodeReportTest(t)
	first := EquipCodeReport{CommanderID: 2, ShipGroupID: 3, ShareID: 4, ReportType: 2, ReportDay: 11}
	second := EquipCodeReport{CommanderID: 2, ShipGroupID: 3, ShareID: 4, ReportType: 1, ReportDay: 11}
	if err := GormDB.Create(&first).Error; err != nil {
		t.Fatalf("create report failed: %v", err)
	}
	if err := GormDB.Create(&second).Error; err == nil {
		t.Fatalf("expected duplicate report insert to fail")
	}
}
