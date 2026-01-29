package connection

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T, models ...any) *gorm.DB {
	t.Helper()
	name := strings.ReplaceAll(t.Name(), "/", "_")
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", name)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{PrepareStmt: true})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	if len(models) > 0 {
		if err := db.AutoMigrate(models...); err != nil {
			t.Fatalf("failed to migrate: %v", err)
		}
	}
	return db
}

func withTestDB(t *testing.T, models ...any) {
	t.Helper()
	originalDB := orm.GormDB
	orm.GormDB = newTestDB(t, models...)
	t.Cleanup(func() {
		orm.GormDB = originalDB
	})
}
