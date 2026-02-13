package orm

import (
	"fmt"
	"strings"
	"testing"
)

func TestQualifiedTable_NoDB(t *testing.T) {
	if got := QualifiedTable("account_roles"); got != "account_roles" {
		t.Fatalf("expected account_roles, got %q", got)
	}
}

func TestQualifiedTable_WithPrefix(t *testing.T) {
	if got := QualifiedTable("account_roles"); got != "account_roles" {
		t.Fatalf("expected account_roles, got %q", got)
	}

	sql := fmt.Sprintf(
		"SELECT roles.name FROM %s AS account_roles JOIN %s AS roles ON roles.id = account_roles.role_id WHERE account_roles.account_id = $1 ORDER BY roles.name asc",
		QualifiedTable("account_roles"),
		QualifiedTable("roles"),
	)

	if !strings.Contains(sql, "account_roles") {
		t.Fatalf("expected query to include qualified account_roles table, got SQL: %s", sql)
	}
}
