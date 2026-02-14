package handlers

import (
	"context"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
)

func execTestSQL(t *testing.T, query string, args ...any) {
	t.Helper()
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), query, args...); err != nil {
		t.Fatalf("exec sql failed: %v", err)
	}
}

func queryInt64TestSQL(t *testing.T, query string, args ...any) int64 {
	t.Helper()
	var value int64
	if err := db.DefaultStore.Pool.QueryRow(context.Background(), query, args...).Scan(&value); err != nil {
		t.Fatalf("query row failed: %v", err)
	}
	return value
}

func queryUint32TestSQL(t *testing.T, query string, args ...any) uint32 {
	t.Helper()
	var value uint32
	if err := db.DefaultStore.Pool.QueryRow(context.Background(), query, args...).Scan(&value); err != nil {
		t.Fatalf("query row failed: %v", err)
	}
	return value
}
