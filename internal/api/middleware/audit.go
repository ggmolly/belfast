package middleware

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/orm"
)

const (
	auditActionKey   = "audit.action"
	auditMetadataKey = "audit.metadata"
	authzKeyKey      = "authz.key"
	authzOpKey       = "authz.op"
)

func SetAuditAction(ctx iris.Context, action string) {
	if action == "" {
		return
	}
	ctx.Values().Set(auditActionKey, action)
}

func AddAuditMetadata(ctx iris.Context, key string, value interface{}) {
	if key == "" {
		return
	}
	meta, _ := ctx.Values().Get(auditMetadataKey).(map[string]interface{})
	if meta == nil {
		meta = map[string]interface{}{}
	}
	meta[key] = value
	ctx.Values().Set(auditMetadataKey, meta)
}

func Audit() iris.Handler {
	return func(ctx iris.Context) {
		method := ctx.Method()
		if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
			ctx.Next()
			return
		}
		start := time.Now()
		ctx.Next()

		status := ctx.GetStatusCode()
		path := ctx.Path()
		var actorAccountID *string
		var actorCommanderID *uint32
		if account, ok := GetAccount(ctx); ok {
			actorAccountID = &account.ID
			if account.CommanderID != nil {
				actorCommanderID = account.CommanderID
			}
		}

		permissionKey, _ := ctx.Values().Get(authzKeyKey).(string)
		permissionOp, _ := ctx.Values().Get(authzOpKey).(string)
		var permissionKeyPtr *string
		var permissionOpPtr *string
		if permissionKey != "" {
			permissionKeyPtr = &permissionKey
		}
		if permissionOp != "" {
			permissionOpPtr = &permissionOp
		}

		action, _ := ctx.Values().Get(auditActionKey).(string)
		meta, _ := ctx.Values().Get(auditMetadataKey).(map[string]interface{})
		if meta == nil {
			meta = map[string]interface{}{}
		}
		meta["latency_ms"] = time.Since(start).Milliseconds()
		payload, _ := json.Marshal(meta)

		entry := orm.AuditLog{
			ID:               uuid.NewString(),
			ActorAccountID:   actorAccountID,
			ActorCommanderID: actorCommanderID,
			Method:           method,
			Path:             path,
			StatusCode:       status,
			PermissionKey:    permissionKeyPtr,
			PermissionOp:     permissionOpPtr,
			Action:           action,
			Metadata:         payload,
			CreatedAt:        time.Now().UTC(),
		}
		_ = orm.GormDB.Create(&entry).Error
	}
}
