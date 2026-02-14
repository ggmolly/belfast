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

func TestSplitSQLStatements(t *testing.T) {
	input := `
-- comment ; should not split
DELETE FROM fleets WHERE commander_id = 1;
CREATE UNIQUE INDEX CONCURRENTLY idx_fleets ON fleets (commander_id, game_id);
DO $$
BEGIN
  PERFORM 1;
END
$$;
`

	statements := splitSQLStatements(input)
	if len(statements) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(statements))
	}

	if statements[0] != "-- comment ; should not split\nDELETE FROM fleets WHERE commander_id = 1" {
		t.Fatalf("unexpected first statement: %q", statements[0])
	}

	if statements[1] != "CREATE UNIQUE INDEX CONCURRENTLY idx_fleets ON fleets (commander_id, game_id)" {
		t.Fatalf("unexpected second statement: %q", statements[1])
	}

	if statements[2] != "DO $$\nBEGIN\n  PERFORM 1;\nEND\n$$" {
		t.Fatalf("unexpected third statement: %q", statements[2])
	}
}
