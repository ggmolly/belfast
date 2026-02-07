package orm

import (
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

const DefaultUserPermissionPolicyID = "user-default"

func LoadUserPermissionPolicy() (UserPermissionPolicy, error) {
	var policy UserPermissionPolicy
	if err := GormDB.First(&policy, "id = ?", DefaultUserPermissionPolicyID).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return UserPermissionPolicy{}, err
		}
		payload, _ := json.Marshal([]string{})
		policy = UserPermissionPolicy{
			ID:        DefaultUserPermissionPolicyID,
			Actions:   payload,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		if err := GormDB.Create(&policy).Error; err != nil {
			return UserPermissionPolicy{}, err
		}
	}
	return policy, nil
}

func UpdateUserPermissionPolicy(actions []string, updatedBy *string) (UserPermissionPolicy, error) {
	policy, err := LoadUserPermissionPolicy()
	if err != nil {
		return UserPermissionPolicy{}, err
	}
	payload, err := json.Marshal(actions)
	if err != nil {
		return UserPermissionPolicy{}, err
	}
	now := time.Now().UTC()
	updates := map[string]interface{}{
		"actions":    payload,
		"updated_by": updatedBy,
		"updated_at": now,
	}
	if err := GormDB.Model(&UserPermissionPolicy{}).Where("id = ?", policy.ID).Updates(updates).Error; err != nil {
		return UserPermissionPolicy{}, err
	}
	policy.Actions = payload
	policy.UpdatedBy = updatedBy
	policy.UpdatedAt = now
	return policy, nil
}

func DecodeUserPermissionActions(policy UserPermissionPolicy) ([]string, error) {
	if len(policy.Actions) == 0 {
		return []string{}, nil
	}
	var actions []string
	if err := json.Unmarshal(policy.Actions, &actions); err != nil {
		return nil, err
	}
	return actions, nil
}

func UserPermissionAllowed(action string) (bool, error) {
	policy, err := LoadUserPermissionPolicy()
	if err != nil {
		return false, err
	}
	actions, err := DecodeUserPermissionActions(policy)
	if err != nil {
		return false, err
	}
	for _, entry := range actions {
		if entry == action {
			return true, nil
		}
	}
	return false, nil
}
