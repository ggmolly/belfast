package auth

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/ggmolly/belfast/internal/orm"
)

func LogAudit(action string, actorUserID *string, targetUserID *string, metadata map[string]interface{}) {
	entry := orm.AdminAuditLog{
		ID:           uuid.NewString(),
		ActorUserID:  actorUserID,
		Action:       action,
		TargetUserID: targetUserID,
		CreatedAt:    time.Now().UTC(),
	}
	if metadata != nil {
		if payload, err := json.Marshal(metadata); err == nil {
			entry.Metadata = payload
		}
	}
	_ = orm.GormDB.Create(&entry).Error
}

func LogUserAudit(action string, actorUserID *string, targetCommanderID *uint32, metadata map[string]interface{}) {
	entry := orm.UserAuditLog{
		ID:                uuid.NewString(),
		ActorUserID:       actorUserID,
		Action:            action,
		TargetCommanderID: targetCommanderID,
		CreatedAt:         time.Now().UTC(),
	}
	if metadata != nil {
		if payload, err := json.Marshal(metadata); err == nil {
			entry.Metadata = payload
		}
	}
	_ = orm.GormDB.Create(&entry).Error
}
