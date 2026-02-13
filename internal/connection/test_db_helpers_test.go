package connection

import (
	"context"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
)

func withTestDB(t *testing.T, models ...any) {
	t.Helper()
	t.Setenv("MODE", "test")
	orm.InitDatabase()
	_ = models
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `
TRUNCATE TABLE
	fleets,
	owned_resources,
	commander_items,
	owned_ships,
	commanders,
	yostarus_maps
RESTART IDENTITY CASCADE
`); err != nil {
		t.Fatalf("failed to reset test tables: %v", err)
	}
}
