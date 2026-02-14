package connection

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
)

func withTestDB(t *testing.T, models ...any) {
	t.Helper()
	t.Setenv("MODE", "test")
	orm.InitDatabase()
	_ = models
	rows, err := db.DefaultStore.Pool.Query(context.Background(), `
SELECT tablename
FROM pg_tables
WHERE schemaname = current_schema()
ORDER BY tablename
`)
	if err != nil {
		t.Fatalf("failed to list test tables: %v", err)
	}
	defer rows.Close()

	excluded := map[string]struct{}{
		"equipments":        {},
		"items":             {},
		"resources":         {},
		"roles":             {},
		"permissions":       {},
		"schema_migrations": {},
		"role_permissions":  {},
		"ships":             {},
	}

	tables := make([]string, 0, 64)
	for rows.Next() {
		var table string
		if scanErr := rows.Scan(&table); scanErr != nil {
			t.Fatalf("failed to scan table name: %v", scanErr)
		}
		if _, skip := excluded[table]; skip {
			continue
		}
		tables = append(tables, fmt.Sprintf(`"%s"`, strings.ReplaceAll(table, `"`, `""`)))
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("failed to iterate table names: %v", err)
	}
	if len(tables) == 0 {
		return
	}

	statement := "TRUNCATE TABLE " + strings.Join(tables, ", ") + " RESTART IDENTITY CASCADE"
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), statement); err != nil {
		t.Fatalf("failed to reset test tables: %v", err)
	}
}
