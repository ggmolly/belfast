package orm

import (
	"context"
	"sync"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
)

var equipCodeShareTestOnce sync.Once

func initEquipCodeShareTest(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	equipCodeShareTestOnce.Do(func() {
		InitDatabase()
	})
	clearTable(t, &EquipCodeShare{})
}

func TestEquipCodeShareCreate(t *testing.T) {
	initEquipCodeShareTest(t)
	share := EquipCodeShare{CommanderID: 1, ShipGroupID: 2, ShareDay: 10}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO equip_code_shares (commander_id, ship_group_id, share_day) VALUES ($1, $2, $3)`, int64(share.CommanderID), int64(share.ShipGroupID), int64(share.ShareDay)); err != nil {
		t.Fatalf("create share failed: %v", err)
	}
	var stored EquipCodeShare
	if err := db.DefaultStore.Pool.QueryRow(context.Background(), `SELECT commander_id, ship_group_id, share_day FROM equip_code_shares WHERE commander_id = $1 AND ship_group_id = $2 AND share_day = $3`, int64(1), int64(2), int64(10)).Scan(&stored.CommanderID, &stored.ShipGroupID, &stored.ShareDay); err != nil {
		t.Fatalf("fetch share failed: %v", err)
	}
	if stored.CommanderID != 1 {
		t.Fatalf("expected commander_id 1, got %d", stored.CommanderID)
	}
}

func TestEquipCodeShareDedupeIndex(t *testing.T) {
	initEquipCodeShareTest(t)
	first := EquipCodeShare{CommanderID: 2, ShipGroupID: 3, ShareDay: 11}
	second := EquipCodeShare{CommanderID: 2, ShipGroupID: 3, ShareDay: 11}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO equip_code_shares (commander_id, ship_group_id, share_day) VALUES ($1, $2, $3)`, int64(first.CommanderID), int64(first.ShipGroupID), int64(first.ShareDay)); err != nil {
		t.Fatalf("create share failed: %v", err)
	}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO equip_code_shares (commander_id, ship_group_id, share_day) VALUES ($1, $2, $3)`, int64(second.CommanderID), int64(second.ShipGroupID), int64(second.ShareDay)); err == nil {
		t.Fatalf("expected duplicate share insert to fail")
	}
}
