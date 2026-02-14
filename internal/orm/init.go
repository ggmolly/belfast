package orm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/db"
)

var (
	initOnce sync.Once
	initErr  error
)

// InitDatabase initializes the Postgres/sqlc store for the current process.
//
// This remains as a compatibility shim for tests and legacy callers.
// Production startup should prefer internal/db bootstrap directly.
func InitDatabase() bool {
	didInit := false
	initOnce.Do(func() {
		didInit = true
		dsn := strings.TrimSpace(os.Getenv("BELFAST_TEST_POSTGRES_DSN"))
		if dsn == "" {
			dsn = strings.TrimSpace(os.Getenv("TEST_DATABASE_DSN"))
		}
		if dsn == "" {
			cfg, cfgErr := loadServerConfig()
			if cfgErr != nil {
				initErr = fmt.Errorf("missing Postgres DSN from env or server.toml: %w", cfgErr)
				return
			}
			dsn = strings.TrimSpace(cfg.DB.DSN)
		}
		if dsn == "" {
			initErr = fmt.Errorf("missing Postgres DSN; set BELFAST_TEST_POSTGRES_DSN, TEST_DATABASE_DSN, or server.toml [database].dsn")
			return
		}
		schemaName := "belfast_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
		_, initErr = db.InitDefaultStore(context.Background(), dsn, schemaName)
		if initErr != nil {
			return
		}
		if strings.EqualFold(strings.TrimSpace(os.Getenv("MODE")), "test") {
			initErr = applyTestDatabaseCompatibility()
			if initErr != nil {
				return
			}
		}

	})
	if initErr != nil {
		panic(initErr.Error())
	}
	return didInit
}

func loadServerConfig() (config.Config, error) {
	const configName = "server.toml"
	startDir, err := os.Getwd()
	if err != nil {
		return config.Config{}, fmt.Errorf("failed to get working directory: %w", err)
	}

	dir := filepath.Clean(startDir)
	for {
		cfgPath := filepath.Join(dir, configName)
		_, statErr := os.Stat(cfgPath)
		if statErr == nil {
			cfg, loadErr := config.Load(cfgPath)
			if loadErr != nil {
				return config.Config{}, fmt.Errorf("failed to load %s: %w", cfgPath, loadErr)
			}
			return cfg, nil
		}
		if !errors.Is(statErr, os.ErrNotExist) {
			return config.Config{}, statErr
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return config.Config{}, fmt.Errorf("config file missing: %s", configName)
		}
		dir = parent
	}
}

func applyTestDatabaseCompatibility() error {
	ctx := context.Background()
	compatSQL := []string{
		`DO $$
DECLARE r record;
BEGIN
  FOR r IN
    SELECT conname, conrelid::regclass::text AS table_name
    FROM pg_constraint
    WHERE contype = 'f'
      AND confrelid = format('%I.commanders', current_schema())::regclass
  LOOP
    EXECUTE format('ALTER TABLE %s DROP CONSTRAINT IF EXISTS %I', r.table_name, r.conname);
  END LOOP;
END $$`,
		`ALTER TABLE commander_items DROP CONSTRAINT IF EXISTS commander_items_item_id_fkey`,
		`ALTER TABLE commander_misc_items DROP CONSTRAINT IF EXISTS commander_misc_items_item_id_fkey`,
		`ALTER TABLE owned_resources DROP CONSTRAINT IF EXISTS owned_resources_resource_id_fkey`,
		`ALTER TABLE owned_ships DROP CONSTRAINT IF EXISTS owned_ships_ship_id_fkey`,
		`ALTER TABLE owned_ship_shadow_skins DROP CONSTRAINT IF EXISTS owned_ship_shadow_skins_ship_id_fkey`,
		`ALTER TABLE builds DROP CONSTRAINT IF EXISTS builds_ship_id_fkey`,
		`ALTER TABLE random_flag_ships DROP CONSTRAINT IF EXISTS random_flag_ships_ship_id_fkey`,
		`ALTER TABLE global_skin_restrictions DROP CONSTRAINT IF EXISTS global_skin_restrictions_skin_id_fkey`,
		`ALTER TABLE global_skin_restriction_windows DROP CONSTRAINT IF EXISTS global_skin_restriction_windows_skin_id_fkey`,
		`CREATE SEQUENCE IF NOT EXISTS owned_spweapons_id_seq`,
		`ALTER TABLE owned_spweapons ALTER COLUMN id SET DEFAULT nextval('owned_spweapons_id_seq')`,
		`ALTER TABLE equipments ALTER COLUMN destroy_gold SET DEFAULT 0`,
		`ALTER TABLE equipments ALTER COLUMN equip_limit SET DEFAULT 0`,
		`ALTER TABLE equipments ALTER COLUMN "group" SET DEFAULT 0`,
		`ALTER TABLE equipments ALTER COLUMN important SET DEFAULT 0`,
		`ALTER TABLE equipments ALTER COLUMN level SET DEFAULT 0`,
		`ALTER TABLE equipments ALTER COLUMN next SET DEFAULT 0`,
		`ALTER TABLE equipments ALTER COLUMN prev SET DEFAULT 0`,
		`ALTER TABLE equipments ALTER COLUMN restore_gold SET DEFAULT 0`,
		`ALTER TABLE equipments ALTER COLUMN trans_use_gold SET DEFAULT 0`,
		`ALTER TABLE equipments ALTER COLUMN type SET DEFAULT 0`,
		`ALTER TABLE equipments ALTER COLUMN destroy_item SET DEFAULT '[]'::jsonb`,
		`ALTER TABLE equipments ALTER COLUMN restore_item SET DEFAULT '[]'::jsonb`,
		`ALTER TABLE equipments ALTER COLUMN ship_type_forbidden SET DEFAULT '[]'::jsonb`,
		`ALTER TABLE equipments ALTER COLUMN trans_use_item SET DEFAULT '[]'::jsonb`,
		`ALTER TABLE equipments ALTER COLUMN upgrade_formula_id SET DEFAULT '[]'::jsonb`,
	}
	for _, statement := range compatSQL {
		if _, err := db.DefaultStore.Pool.Exec(ctx, statement); err != nil {
			return err
		}
	}
	if _, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time)
VALUES
	(202124, 'Belfast', 'Belfast', 5, 6, 2, 1, 0),
	(106011, 'Long Island', 'Long Island', 2, 2, 7, 1, 0),
	(201211, 'Starter', 'Starter', 2, 2, 1, 1, 0)
ON CONFLICT (template_id) DO NOTHING`); err != nil {
		return err
	}
	return nil
}
