package orm

import (
	"context"
	"sync"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
)

var equipCodeLikeTestOnce sync.Once

func initEquipCodeLikeTest(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	equipCodeLikeTestOnce.Do(func() {
		InitDatabase()
	})
	clearTable(t, &EquipCodeLike{})
}

func TestEquipCodeLikeCreate(t *testing.T) {
	initEquipCodeLikeTest(t)
	like := EquipCodeLike{CommanderID: 1, ShipGroupID: 2, ShareID: 3, LikeDay: 10}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO equip_code_likes (commander_id, ship_group_id, share_id, like_day) VALUES ($1, $2, $3, $4)`, int64(like.CommanderID), int64(like.ShipGroupID), int64(like.ShareID), int64(like.LikeDay)); err != nil {
		t.Fatalf("create like failed: %v", err)
	}

	var stored EquipCodeLike
	if err := db.DefaultStore.Pool.QueryRow(context.Background(), `SELECT commander_id, ship_group_id, share_id, like_day FROM equip_code_likes WHERE commander_id = $1 AND share_id = $2`, int64(1), int64(3)).Scan(&stored.CommanderID, &stored.ShipGroupID, &stored.ShareID, &stored.LikeDay); err != nil {
		t.Fatalf("fetch like failed: %v", err)
	}
	if stored.ShipGroupID != 2 {
		t.Fatalf("expected shipgroup 2, got %d", stored.ShipGroupID)
	}
}

func TestEquipCodeLikeDedupeIndex(t *testing.T) {
	initEquipCodeLikeTest(t)
	first := EquipCodeLike{CommanderID: 2, ShipGroupID: 3, ShareID: 4, LikeDay: 11}
	second := EquipCodeLike{CommanderID: 2, ShipGroupID: 3, ShareID: 4, LikeDay: 11}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO equip_code_likes (commander_id, ship_group_id, share_id, like_day) VALUES ($1, $2, $3, $4)`, int64(first.CommanderID), int64(first.ShipGroupID), int64(first.ShareID), int64(first.LikeDay)); err != nil {
		t.Fatalf("create like failed: %v", err)
	}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO equip_code_likes (commander_id, ship_group_id, share_id, like_day) VALUES ($1, $2, $3, $4)`, int64(second.CommanderID), int64(second.ShipGroupID), int64(second.ShareID), int64(second.LikeDay)); err == nil {
		t.Fatalf("expected duplicate like insert to fail")
	}
}

func TestEquipCodeLikeDedupeAllowsDifferentShipGroup(t *testing.T) {
	initEquipCodeLikeTest(t)
	first := EquipCodeLike{CommanderID: 2, ShipGroupID: 3, ShareID: 4, LikeDay: 11}
	second := EquipCodeLike{CommanderID: 2, ShipGroupID: 999, ShareID: 4, LikeDay: 11}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO equip_code_likes (commander_id, ship_group_id, share_id, like_day) VALUES ($1, $2, $3, $4)`, int64(first.CommanderID), int64(first.ShipGroupID), int64(first.ShareID), int64(first.LikeDay)); err != nil {
		t.Fatalf("create like failed: %v", err)
	}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO equip_code_likes (commander_id, ship_group_id, share_id, like_day) VALUES ($1, $2, $3, $4)`, int64(second.CommanderID), int64(second.ShipGroupID), int64(second.ShareID), int64(second.LikeDay)); err != nil {
		t.Fatalf("create like with different shipgroup failed: %v", err)
	}
}
