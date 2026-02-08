package orm

import (
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestQualifiedTable_NoDB(t *testing.T) {
	old := GormDB
	GormDB = nil
	t.Cleanup(func() { GormDB = old })

	if got := QualifiedTable("account_roles"); got != "account_roles" {
		t.Fatalf("expected account_roles, got %q", got)
	}
}

func TestQualifiedTable_WithPrefix(t *testing.T) {
	old := GormDB
	t.Cleanup(func() { GormDB = old })

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		DryRun: true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: "belfast.",
		},
	})
	if err != nil {
		t.Fatalf("open gorm: %v", err)
	}
	GormDB = db

	if got := QualifiedTable("account_roles"); got != "belfast.account_roles" {
		t.Fatalf("expected belfast.account_roles, got %q", got)
	}

	sql := GormDB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		var names []string
		return tx.Table(QualifiedTable("account_roles")+" AS account_roles").
			Select("roles.name").
			Joins("JOIN "+QualifiedTable("roles")+" AS roles ON roles.id = account_roles.role_id").
			Where("account_roles.account_id = ?", "abc").
			Order("roles.name asc").
			Scan(&names)
	})

	if !(strings.Contains(sql, "belfast.account_roles") ||
		strings.Contains(sql, "\"belfast\".\"account_roles\"") ||
		strings.Contains(sql, "`belfast`.`account_roles`")) {
		t.Fatalf("expected query to include qualified account_roles table, got SQL: %s", sql)
	}
}
