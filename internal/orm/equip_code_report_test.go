package orm

import (
	"context"
	"sync"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
)

var equipCodeReportTestOnce sync.Once

func initEquipCodeReportTest(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	equipCodeReportTestOnce.Do(func() {
		InitDatabase()
	})
	clearTable(t, &EquipCodeReport{})
}

func TestEquipCodeReportCreate(t *testing.T) {
	initEquipCodeReportTest(t)
	report := EquipCodeReport{CommanderID: 1, ShipGroupID: 2, ShareID: 3, ReportType: 1, ReportDay: 10}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO equip_code_reports (commander_id, ship_group_id, share_id, report_type, report_day) VALUES ($1, $2, $3, $4, $5)`, int64(report.CommanderID), int64(report.ShipGroupID), int64(report.ShareID), int64(report.ReportType), int64(report.ReportDay)); err != nil {
		t.Fatalf("create report failed: %v", err)
	}
	var stored EquipCodeReport
	if err := db.DefaultStore.Pool.QueryRow(context.Background(), `SELECT commander_id, ship_group_id, share_id, report_type, report_day FROM equip_code_reports WHERE commander_id = $1 AND share_id = $2`, int64(1), int64(3)).Scan(&stored.CommanderID, &stored.ShipGroupID, &stored.ShareID, &stored.ReportType, &stored.ReportDay); err != nil {
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
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO equip_code_reports (commander_id, ship_group_id, share_id, report_type, report_day) VALUES ($1, $2, $3, $4, $5)`, int64(first.CommanderID), int64(first.ShipGroupID), int64(first.ShareID), int64(first.ReportType), int64(first.ReportDay)); err != nil {
		t.Fatalf("create report failed: %v", err)
	}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO equip_code_reports (commander_id, ship_group_id, share_id, report_type, report_day) VALUES ($1, $2, $3, $4, $5)`, int64(second.CommanderID), int64(second.ShipGroupID), int64(second.ShareID), int64(second.ReportType), int64(second.ReportDay)); err == nil {
		t.Fatalf("expected duplicate report insert to fail")
	}
}
