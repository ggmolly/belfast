package orm

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

func CreateAuditLog(entry AuditLog) error {
	if db.DefaultStore == nil {
		return errors.New("db not initialized")
	}
	ctx := context.Background()
	return db.DefaultStore.Queries.CreateAuditLog(ctx, gen.CreateAuditLogParams{
		ID:               entry.ID,
		ActorAccountID:   pgTextFromPtr(entry.ActorAccountID),
		ActorCommanderID: pgInt8FromUint32Ptr(entry.ActorCommanderID),
		Method:           entry.Method,
		Path:             entry.Path,
		StatusCode:       int32(entry.StatusCode),
		PermissionKey:    pgTextFromPtr(entry.PermissionKey),
		PermissionOp:     pgTextFromPtr(entry.PermissionOp),
		Action:           pgtype.Text{String: entry.Action, Valid: true},
		Metadata:         entry.Metadata,
		CreatedAt:        pgTimestamptz(entry.CreatedAt),
	})
}
