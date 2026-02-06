package auth

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

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
	_ = orm.GormDB.Create(&entry).Error
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
	_ = orm.GormDB.Create(&entry).Error
}
