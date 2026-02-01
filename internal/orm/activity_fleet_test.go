package orm

import (
	"os"
	"testing"

	"gorm.io/gorm"
)

func TestActivityFleetGroupRoundTrip(t *testing.T) {
	os.Setenv("MODE", "test")
	InitDatabase()
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&ActivityFleet{}).Error; err != nil {
		t.Fatalf("failed to clear activity fleets: %v", err)
	}
	groups := ActivityFleetGroupList{
		{
			ID:       1,
			ShipList: []uint32{10, 11},
			Commanders: []ActivityFleetCommander{
				{Pos: 1, ID: 99},
				{Pos: 2, ID: 100},
			},
		},
	}
	if err := SaveActivityFleetGroups(7, 42, groups); err != nil {
		t.Fatalf("save activity fleet groups failed: %v", err)
	}
	loaded, found, err := LoadActivityFleetGroups(7, 42)
	if err != nil {
		t.Fatalf("load activity fleet groups failed: %v", err)
	}
	if !found {
		t.Fatalf("expected activity fleet groups to be found")
	}
	if len(loaded) != 1 {
		t.Fatalf("expected 1 group, got %d", len(loaded))
	}
	if loaded[0].ID != 1 {
		t.Fatalf("expected group id 1, got %d", loaded[0].ID)
	}
	if len(loaded[0].ShipList) != 2 || loaded[0].ShipList[0] != 10 || loaded[0].ShipList[1] != 11 {
		t.Fatalf("unexpected ship list: %v", loaded[0].ShipList)
	}
	if len(loaded[0].Commanders) != 2 || loaded[0].Commanders[0].ID != 99 || loaded[0].Commanders[1].ID != 100 {
		t.Fatalf("unexpected commanders: %v", loaded[0].Commanders)
	}
}
