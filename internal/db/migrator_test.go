package db

import "testing"

func TestLoadEmbeddedMigrations(t *testing.T) {
	migrations, err := LoadEmbeddedMigrations()
	if err != nil {
		t.Fatalf("LoadEmbeddedMigrations() err = %v", err)
	}
	if len(migrations) == 0 {
		t.Fatalf("expected at least one embedded migration")
	}
	if migrations[0].Version != 1 {
		t.Fatalf("expected first migration version 1, got %d", migrations[0].Version)
	}
}
