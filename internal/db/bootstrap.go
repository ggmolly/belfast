package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

func OpenPostgresSQLDB(ctx context.Context, dsn string) (*sql.DB, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, fmt.Errorf("postgres dsn is required")
	}
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	db := stdlib.OpenDB(*cfg)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func InitDefaultStore(ctx context.Context, dsn string, schemaName string) (*Store, error) {
	sqlDB, err := OpenPostgresSQLDB(ctx, dsn)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = sqlDB.Close()
	}()

	if schema := strings.TrimSpace(schemaName); schema != "" {
		if _, err := sqlDB.ExecContext(ctx, `CREATE SCHEMA IF NOT EXISTS `+quoteIdent(schema)); err != nil {
			return nil, err
		}
	}
	if err := RunMigrations(ctx, sqlDB, MigratorOptions{SchemaName: schemaName}); err != nil {
		return nil, err
	}

	pool, err := OpenPostgresPool(ctx, dsn, schemaName)
	if err != nil {
		return nil, err
	}
	store := NewStore(pool)
	DefaultStore = store
	return store, nil
}

func HasGameData(ctx context.Context, store *Store) (bool, error) {
	// Items are always imported as part of initial game data.
	// Using EXISTS keeps this cheap even on large datasets.
	row := store.Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM items LIMIT 1)::bool`)
	var ok bool
	if err := row.Scan(&ok); err != nil {
		return false, err
	}
	return ok, nil
}
