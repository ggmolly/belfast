package orm

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/db"
)

var accountAuthQueriesOnce sync.Once

func initAccountAuthQueriesDB(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	accountAuthQueriesOnce.Do(func() {
		InitDatabase()
	})
}

func clearAccountAuthQueryTables(t *testing.T) {
	t.Helper()
	tables := []string{
		"account_permission_overrides",
		"account_roles",
		"web_authn_credentials",
		"auth_challenges",
		"sessions",
		"audit_logs",
		"accounts",
	}
	for _, table := range tables {
		if _, err := db.DefaultStore.Pool.Exec(t.Context(), "DELETE FROM "+table); err != nil {
			t.Fatalf("clear table %s: %v", table, err)
		}
	}
}

func TestAccountAuthQueriesNotFoundOnZeroRows(t *testing.T) {
	initAccountAuthQueriesDB(t)
	clearAccountAuthQueryTables(t)

	now := time.Now().UTC()

	err := UpdateAccountUsername("missing-account", "admin", "admin", now)
	if !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("expected UpdateAccountUsername to return db.ErrNotFound, got %v", err)
	}

	err = UpdateAccountDisabledAt("missing-account", &now, now)
	if !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("expected UpdateAccountDisabledAt to return db.ErrNotFound, got %v", err)
	}

	err = UpdateAccountPassword("missing-account", "hash", "argon2id", now, now)
	if !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("expected UpdateAccountPassword to return db.ErrNotFound, got %v", err)
	}

	err = DeleteAccountByID("missing-account")
	if !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("expected DeleteAccountByID to return db.ErrNotFound, got %v", err)
	}
}
