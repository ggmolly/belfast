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
	for i := 1; i < len(migrations); i++ {
		if migrations[i].Version <= migrations[i-1].Version {
			t.Fatalf("expected strictly increasing versions, got %d then %d", migrations[i-1].Version, migrations[i].Version)
		}
		if migrations[i].Filename == "" {
			t.Fatalf("expected non-empty migration filename for version %d", migrations[i].Version)
		}
	}
}

func TestHasNoTransactionDirective(t *testing.T) {
	withDirective := "-- +migrate NoTransaction\nCREATE INDEX CONCURRENTLY idx ON t (id);"
	if !hasNoTransactionDirective(withDirective) {
		t.Fatalf("expected NoTransaction directive to be detected")
	}

	withoutDirective := "-- comment\nCREATE TABLE example(id bigint);"
	if hasNoTransactionDirective(withoutDirective) {
		t.Fatalf("did not expect NoTransaction directive to be detected")
	}
}
