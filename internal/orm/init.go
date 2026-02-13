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
