package orm

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
)

func resetInitStateForTest(t *testing.T) {
	t.Helper()
	initOnce = sync.Once{}
	initErr = nil
	if db.DefaultStore != nil && db.DefaultStore.Pool != nil {
		db.DefaultStore.Pool.Close()
		db.DefaultStore = nil
	}
}

func resolveInitDatabaseDSN(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("BELFAST_TEST_POSTGRES_DSN")
	if dsn == "" {
		dsn = os.Getenv("TEST_DATABASE_DSN")
	}
	if dsn != "" {
		return dsn
	}
	cfg, err := loadServerConfig()
	if err != nil {
		t.Fatalf("failed to resolve postgres dsn for tests: %v", err)
	}
	if cfg.DB.DSN == "" {
		t.Fatal("server config database dsn is empty")
	}
	return cfg.DB.DSN
}

func TestInitDatabasePanicsWithoutDSN(t *testing.T) {
	t.Cleanup(func() {
		resetInitStateForTest(t)
	})

	resetInitStateForTest(t)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(cwd) }()
	workdir := t.TempDir()
	if err := os.Chdir(workdir); err != nil {
		t.Fatalf("failed to switch to temp dir: %v", err)
	}
	t.Setenv("BELFAST_TEST_POSTGRES_DSN", "")
	t.Setenv("TEST_DATABASE_DSN", "")
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected InitDatabase to panic without DSN")
		}
	}()
	InitDatabase()
}

func TestInitDatabaseSuccessAndIdempotent(t *testing.T) {
	t.Cleanup(func() {
		resetInitStateForTest(t)
	})

	resetInitStateForTest(t)
	t.Setenv("MODE", "test")
	t.Setenv("BELFAST_TEST_POSTGRES_DSN", resolveInitDatabaseDSN(t))
	t.Setenv("TEST_DATABASE_DSN", "")

	if didInit := InitDatabase(); !didInit {
		t.Fatalf("expected first InitDatabase call to initialize store")
	}
	if db.DefaultStore == nil || db.DefaultStore.Pool == nil {
		t.Fatalf("expected initialized default store")
	}

	pool := db.DefaultStore.Pool
	if didInit := InitDatabase(); didInit {
		t.Fatalf("expected second InitDatabase call to be idempotent")
	}
	if db.DefaultStore == nil || db.DefaultStore.Pool != pool {
		t.Fatalf("expected store pool to remain unchanged on idempotent init")
	}

	var schemaName string
	if err := db.DefaultStore.Pool.QueryRow(context.Background(), `SELECT current_schema()`).Scan(&schemaName); err != nil {
		t.Fatalf("failed to query current schema: %v", err)
	}
	if schemaName == "" {
		t.Fatalf("expected non-empty schema")
	}
}
