package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func migratorIntegrationDSN() string {
	dsn := strings.TrimSpace(os.Getenv("BELFAST_TEST_POSTGRES_DSN"))
	if dsn == "" {
		dsn = strings.TrimSpace(os.Getenv("TEST_DATABASE_DSN"))
	}
	if dsn == "" {
		dsn = "postgres://belfast:belfast@localhost:5432/belfast?sslmode=disable"
	}
	return dsn
}

func openMigratorIntegrationDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := OpenPostgresSQLDB(context.Background(), migratorIntegrationDSN())
	if err != nil {
		t.Fatalf("failed to open migrator integration DB: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func newMigratorTestSchema() string {
	return fmt.Sprintf("belfast_migrator_test_%d", time.Now().UnixNano())
}

func TestRunMigrationsIdempotentIntegration(t *testing.T) {
	sqlDB := openMigratorIntegrationDB(t)
	schema := newMigratorTestSchema()
	ctx := context.Background()

	if _, err := sqlDB.ExecContext(ctx, `CREATE SCHEMA IF NOT EXISTS `+quoteIdent(schema)); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	t.Cleanup(func() {
		_, _ = sqlDB.ExecContext(context.Background(), `DROP SCHEMA IF EXISTS `+quoteIdent(schema)+` CASCADE`)
	})

	if err := RunMigrations(ctx, sqlDB, MigratorOptions{SchemaName: schema}); err != nil {
		t.Fatalf("first RunMigrations: %v", err)
	}

	migrations, err := LoadEmbeddedMigrations()
	if err != nil {
		t.Fatalf("LoadEmbeddedMigrations: %v", err)
	}

	var firstCount int
	if err := sqlDB.QueryRowContext(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s`, qualifiedName(schema, "schema_migrations"))).Scan(&firstCount); err != nil {
		t.Fatalf("count schema_migrations after first run: %v", err)
	}
	if firstCount != len(migrations) {
		t.Fatalf("expected %d applied migrations, got %d", len(migrations), firstCount)
	}

	if err := RunMigrations(ctx, sqlDB, MigratorOptions{SchemaName: schema}); err != nil {
		t.Fatalf("second RunMigrations: %v", err)
	}

	var secondCount int
	if err := sqlDB.QueryRowContext(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s`, qualifiedName(schema, "schema_migrations"))).Scan(&secondCount); err != nil {
		t.Fatalf("count schema_migrations after second run: %v", err)
	}
	if secondCount != firstCount {
		t.Fatalf("expected idempotent migration count %d, got %d", firstCount, secondCount)
	}
}

func TestRunMigrationsAdvisoryLockTimeoutIntegration(t *testing.T) {
	sqlDB := openMigratorIntegrationDB(t)
	schema := newMigratorTestSchema()
	ctx := context.Background()

	if _, err := sqlDB.ExecContext(ctx, `CREATE SCHEMA IF NOT EXISTS `+quoteIdent(schema)); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	t.Cleanup(func() {
		_, _ = sqlDB.ExecContext(context.Background(), `DROP SCHEMA IF EXISTS `+quoteIdent(schema)+` CASCADE`)
	})

	lockConn, err := sqlDB.Conn(ctx)
	if err != nil {
		t.Fatalf("open lock connection: %v", err)
	}
	t.Cleanup(func() { _ = lockConn.Close() })

	lockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	lockObjectID := migrationAdvisoryLockObjectID(schema)
	if _, err := lockConn.ExecContext(lockCtx, `SELECT pg_advisory_lock($1, $2)`, migrationAdvisoryLockClassID, lockObjectID); err != nil {
		t.Fatalf("acquire advisory lock: %v", err)
	}
	t.Cleanup(func() {
		unlockCtx, unlockCancel := context.WithTimeout(context.Background(), migrationResetTimeout)
		defer unlockCancel()
		_, _ = lockConn.ExecContext(unlockCtx, `SELECT pg_advisory_unlock($1, $2)`, migrationAdvisoryLockClassID, lockObjectID)
	})

	lockedCtx, lockedCancel := context.WithTimeout(context.Background(), 750*time.Millisecond)
	defer lockedCancel()

	start := time.Now()
	err = RunMigrations(lockedCtx, sqlDB, MigratorOptions{SchemaName: schema})
	if err == nil {
		t.Fatalf("expected advisory lock contention error")
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("expected bounded migration lock acquisition, took %s", elapsed)
	}
	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(strings.ToLower(err.Error()), "timeout") {
		t.Fatalf("expected timeout/deadline error, got %v", err)
	}
}

func TestApplyMigrationFailureDoesNotRecordVersionIntegration(t *testing.T) {
	sqlDB := openMigratorIntegrationDB(t)
	schema := newMigratorTestSchema()
	ctx := context.Background()

	if _, err := sqlDB.ExecContext(ctx, `CREATE SCHEMA IF NOT EXISTS `+quoteIdent(schema)); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	t.Cleanup(func() {
		_, _ = sqlDB.ExecContext(context.Background(), `DROP SCHEMA IF EXISTS `+quoteIdent(schema)+` CASCADE`)
	})

	if err := ensureSchemaMigrationsTable(ctx, sqlDB, schema); err != nil {
		t.Fatalf("ensure schema_migrations: %v", err)
	}

	m := Migration{
		Version:  987654321,
		Name:     "broken_sql",
		Filename: "987654321_broken_sql.sql",
		SQL:      "CREATE TABLE broken_table(",
	}

	if err := applyMigration(ctx, sqlDB, schema, m); err == nil {
		t.Fatalf("expected applyMigration to fail for invalid SQL")
	}

	var recorded int
	if err := sqlDB.QueryRowContext(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE version = $1`, qualifiedName(schema, "schema_migrations")), m.Version).Scan(&recorded); err != nil {
		t.Fatalf("count failed migration record: %v", err)
	}
	if recorded != 0 {
		t.Fatalf("expected failed migration to be unrecorded, got %d", recorded)
	}
}
