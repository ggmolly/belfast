package auth

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
	"github.com/ggmolly/belfast/internal/orm"
)

func LogAudit(action string, actorUserID *string, targetUserID *string, metadata map[string]interface{}) {
	if metadata == nil {
		metadata = map[string]interface{}{}
	}
	if targetUserID != nil {
		metadata["target_account_id"] = *targetUserID
	}
	entry := orm.AuditLog{
		ID:             uuid.NewString(),
		ActorAccountID: actorUserID,
		Method:         "EVENT",
		Path:           "/",
		StatusCode:     0,
		Action:         action,
		CreatedAt:      time.Now().UTC(),
	}
	if metadata != nil {
		if payload, err := json.Marshal(metadata); err == nil {
			entry.Metadata = payload
		}
	}
	if db.DefaultStore == nil {
		return
	}
	ctx := context.Background()
	var actor pgtype.Text
	if entry.ActorAccountID != nil {
		actor = pgtype.Text{String: *entry.ActorAccountID, Valid: true}
	}
	_ = db.DefaultStore.Queries.CreateAuditLog(ctx, gen.CreateAuditLogParams{
		ID:               entry.ID,
		ActorAccountID:   actor,
		ActorCommanderID: pgtype.Int8{},
		Method:           entry.Method,
		Path:             entry.Path,
		StatusCode:       int32(entry.StatusCode),
		PermissionKey:    pgtype.Text{},
		PermissionOp:     pgtype.Text{},
		Action:           pgtype.Text{String: entry.Action, Valid: true},
		Metadata:         entry.Metadata,
		CreatedAt:        pgtype.Timestamptz{Time: entry.CreatedAt, Valid: true},
	})
}

func LogUserAudit(action string, actorUserID *string, targetCommanderID *uint32, metadata map[string]interface{}) {
	if metadata == nil {
		metadata = map[string]interface{}{}
	}
	if targetCommanderID != nil {
		metadata["target_commander_id"] = *targetCommanderID
	}
	entry := orm.AuditLog{
		ID:             uuid.NewString(),
		ActorAccountID: actorUserID,
		Method:         "EVENT",
		Path:           "/",
		StatusCode:     0,
		Action:         action,
		CreatedAt:      time.Now().UTC(),
	}
	if metadata != nil {
		if payload, err := json.Marshal(metadata); err == nil {
			entry.Metadata = payload
		}
	}
	if db.DefaultStore == nil {
		return
	}
	ctx := context.Background()
	var actor pgtype.Text
	if entry.ActorAccountID != nil {
		actor = pgtype.Text{String: *entry.ActorAccountID, Valid: true}
	}
	_ = db.DefaultStore.Queries.CreateAuditLog(ctx, gen.CreateAuditLogParams{
		ID:               entry.ID,
		ActorAccountID:   actor,
		ActorCommanderID: pgtype.Int8{},
		Method:           entry.Method,
		Path:             entry.Path,
		StatusCode:       int32(entry.StatusCode),
		PermissionKey:    pgtype.Text{},
		PermissionOp:     pgtype.Text{},
		Action:           pgtype.Text{String: entry.Action, Valid: true},
		Metadata:         entry.Metadata,
		CreatedAt:        pgtype.Timestamptz{Time: entry.CreatedAt, Valid: true},
	})
}
