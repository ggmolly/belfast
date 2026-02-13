package orm

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

func WithTx(ctx context.Context, fn func(q *gen.Queries) error) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	return db.DefaultStore.WithTx(ctx, fn)
}

func WithPGXTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	return db.DefaultStore.WithPGXTx(ctx, fn)
}
