package answer_test

import (
	"context"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
)

func execAnswerExternalTestSQLT(t *testing.T, query string, args ...any) {
	t.Helper()
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), query, args...); err != nil {
		t.Fatalf("exec sql failed: %v", err)
	}
}

func queryAnswerExternalTestInt64(t *testing.T, query string, args ...any) int64 {
	t.Helper()
	var value int64
	if err := db.DefaultStore.Pool.QueryRow(context.Background(), query, args...).Scan(&value); err != nil {
		t.Fatalf("query row failed: %v", err)
	}
	return value
}
