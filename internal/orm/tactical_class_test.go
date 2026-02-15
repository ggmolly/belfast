package orm

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/ggmolly/belfast/internal/db"
)

func TestCommanderSkillClassLifecycleTx(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &CommanderSkillClass{})
	clearTable(t, &CommanderShipSkill{})
	clearTable(t, &OwnedShip{})
	clearTable(t, &Ship{})
	clearTable(t, &Commander{})

	seedTacticsCommanderShip(t, 8101, 9101, 7101)

	ctx := context.Background()
	entry := CommanderSkillClass{
		CommanderID: 8101,
		RoomID:      1,
		ShipID:      9101,
		SkillPos:    1,
		SkillID:     501,
		StartTime:   10,
		FinishTime:  20,
		Exp:         30,
	}
	err := WithPGXTx(ctx, func(tx pgx.Tx) error {
		return CreateCommanderSkillClassTx(ctx, tx, &entry)
	})
	if err != nil {
		t.Fatalf("create class: %v", err)
	}

	classes, err := ListCommanderSkillClasses(8101)
	if err != nil {
		t.Fatalf("list classes: %v", err)
	}
	if len(classes) != 1 || classes[0].RoomID != 1 || classes[0].ShipID != 9101 {
		t.Fatalf("unexpected classes: %+v", classes)
	}

	err = WithPGXTx(ctx, func(tx pgx.Tx) error {
		locked, err := GetCommanderSkillClassByRoomTx(ctx, tx, 8101, 1)
		if err != nil {
			return err
		}
		if locked.SkillID != 501 {
			t.Fatalf("unexpected skill id %d", locked.SkillID)
		}
		return DeleteCommanderSkillClassTx(ctx, tx, 8101, 1)
	})
	if err != nil {
		t.Fatalf("delete class: %v", err)
	}

	classes, err = ListCommanderSkillClasses(8101)
	if err != nil {
		t.Fatalf("list classes: %v", err)
	}
	if len(classes) != 0 {
		t.Fatalf("expected no classes, got %d", len(classes))
	}
}

func TestCommanderSkillClassConflictOnDuplicateRoom(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &CommanderSkillClass{})
	clearTable(t, &CommanderShipSkill{})
	clearTable(t, &OwnedShip{})
	clearTable(t, &Ship{})
	clearTable(t, &Commander{})

	seedTacticsCommanderShip(t, 8102, 9102, 7102)
	seedTacticsCommanderShip(t, 8102, 9103, 7103)

	ctx := context.Background()
	err := WithPGXTx(ctx, func(tx pgx.Tx) error {
		first := CommanderSkillClass{CommanderID: 8102, RoomID: 1, ShipID: 9102, SkillPos: 1, SkillID: 501, StartTime: 10, FinishTime: 20, Exp: 30}
		if err := CreateCommanderSkillClassTx(ctx, tx, &first); err != nil {
			return err
		}
		second := CommanderSkillClass{CommanderID: 8102, RoomID: 1, ShipID: 9103, SkillPos: 1, SkillID: 502, StartTime: 10, FinishTime: 20, Exp: 30}
		return CreateCommanderSkillClassTx(ctx, tx, &second)
	})
	if !errors.Is(err, ErrSkillClassConflict) {
		t.Fatalf("expected conflict error, got %v", err)
	}
}

func seedTacticsCommanderShip(t *testing.T, commanderID uint32, ownedShipID uint32, templateID uint32) {
	t.Helper()
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commanders (commander_id, account_id, name) VALUES ($1, $2, $3) ON CONFLICT (commander_id) DO NOTHING`, int64(commanderID), int64(1), "Tactics Tester"); err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (template_id) DO NOTHING`, int64(templateID), "Ship", "Ship", int64(2), int64(1), int64(1), int64(1), int64(1)); err != nil {
		t.Fatalf("seed ship: %v", err)
	}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO owned_ships (owner_id, ship_id, id) VALUES ($1, $2, $3)`, int64(commanderID), int64(templateID), int64(ownedShipID)); err != nil {
		t.Fatalf("seed owned ship: %v", err)
	}
}
