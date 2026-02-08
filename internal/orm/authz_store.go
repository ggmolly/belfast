package orm

import (
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/ggmolly/belfast/internal/authz"
)

func ListRoles() ([]Role, error) {
	var roles []Role
	if err := GormDB.Order("name asc").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func ListPermissions() ([]Permission, error) {
	var perms []Permission
	if err := GormDB.Order("key asc").Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

func ListAccountRoleNames(accountID string) ([]string, error) {
	if accountID == "" {
		return []string{}, nil
	}
	var names []string
	if err := GormDB.Table("account_roles").
		Select("roles.name").
		Joins("JOIN roles ON roles.id = account_roles.role_id").
		Where("account_roles.account_id = ?", accountID).
		Order("roles.name asc").
		Scan(&names).Error; err != nil {
		return nil, err
	}
	return names, nil
}

func ReplaceAccountRolesByName(accountID string, roleNames []string) error {
	if accountID == "" {
		return nil
	}
	seen := map[string]struct{}{}
	unique := make([]string, 0, len(roleNames))
	for _, name := range roleNames {
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		unique = append(unique, name)
	}
	return GormDB.Transaction(func(tx *gorm.DB) error {
		var current []string
		if err := tx.Table("account_roles").
			Select("roles.name").
			Joins("JOIN roles ON roles.id = account_roles.role_id").
			Where("account_roles.account_id = ?", accountID).
			Scan(&current).Error; err != nil {
			return err
		}
		currentSet := map[string]struct{}{}
		for _, name := range current {
			currentSet[name] = struct{}{}
		}
		nextSet := map[string]struct{}{}
		for _, name := range unique {
			nextSet[name] = struct{}{}
		}
		if _, hadAdmin := currentSet[authz.RoleAdmin]; hadAdmin {
			if _, willHaveAdmin := nextSet[authz.RoleAdmin]; !willHaveAdmin {
				if err := ensureNotLastRole(tx, authz.RoleAdmin, accountID); err != nil {
					return err
				}
			}
		}

		roleIDs := make([]string, 0, len(unique))
		for _, name := range unique {
			var role Role
			if err := tx.First(&role, "name = ?", name).Error; err != nil {
				return err
			}
			roleIDs = append(roleIDs, role.ID)
		}

		if err := tx.Where("account_id = ?", accountID).Delete(&AccountRole{}).Error; err != nil {
			return err
		}
		now := time.Now().UTC()
		for _, roleID := range roleIDs {
			link := AccountRole{AccountID: accountID, RoleID: roleID, CreatedAt: now}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&link).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

type AccountOverrideEntry struct {
	Key        string
	Mode       string
	Capability authz.Capability
}

func ListAccountOverrides(accountID string) ([]AccountOverrideEntry, error) {
	if accountID == "" {
		return []AccountOverrideEntry{}, nil
	}
	type row struct {
		Key          string
		Mode         string
		CanReadSelf  bool
		CanReadAny   bool
		CanWriteSelf bool
		CanWriteAny  bool
	}
	var rows []row
	if err := GormDB.Table("account_permission_overrides").
		Select("permissions.key as key, account_permission_overrides.mode, account_permission_overrides.can_read_self, account_permission_overrides.can_read_any, account_permission_overrides.can_write_self, account_permission_overrides.can_write_any").
		Joins("JOIN permissions ON permissions.id = account_permission_overrides.permission_id").
		Where("account_permission_overrides.account_id = ?", accountID).
		Order("permissions.key asc").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	entries := make([]AccountOverrideEntry, 0, len(rows))
	for _, r := range rows {
		entries = append(entries, AccountOverrideEntry{
			Key:  r.Key,
			Mode: r.Mode,
			Capability: authz.Capability{
				ReadSelf:  r.CanReadSelf,
				ReadAny:   r.CanReadAny,
				WriteSelf: r.CanWriteSelf,
				WriteAny:  r.CanWriteAny,
			},
		})
	}
	return entries, nil
}

func ReplaceAccountOverrides(accountID string, overrides []AccountOverrideEntry) error {
	if accountID == "" {
		return nil
	}
	seen := map[string]struct{}{}
	unique := make([]AccountOverrideEntry, 0, len(overrides))
	for _, entry := range overrides {
		if entry.Key == "" {
			continue
		}
		if _, ok := seen[entry.Key]; ok {
			continue
		}
		seen[entry.Key] = struct{}{}
		unique = append(unique, entry)
	}
	return GormDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("account_id = ?", accountID).Delete(&AccountPermissionOverride{}).Error; err != nil {
			return err
		}
		now := time.Now().UTC()
		for _, entry := range unique {
			var perm Permission
			if err := tx.First(&perm, "key = ?", entry.Key).Error; err != nil {
				return err
			}
			mode := entry.Mode
			if mode != PermissionOverrideAllow && mode != PermissionOverrideDeny {
				mode = PermissionOverrideAllow
			}
			record := AccountPermissionOverride{
				AccountID:    accountID,
				PermissionID: perm.ID,
				Mode:         mode,
				CanReadSelf:  entry.Capability.ReadSelf,
				CanReadAny:   entry.Capability.ReadAny,
				CanWriteSelf: entry.Capability.WriteSelf,
				CanWriteAny:  entry.Capability.WriteAny,
				UpdatedAt:    now,
			}
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func ensureNotLastRole(tx *gorm.DB, roleName string, excludeAccountID string) error {
	var role Role
	if err := tx.First(&role, "name = ?", roleName).Error; err != nil {
		return err
	}
	var count int64
	query := tx.Model(&AccountRole{}).Where("role_id = ?", role.ID)
	if excludeAccountID != "" {
		query = query.Where("account_id <> ?", excludeAccountID)
	}
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("last role")
	}
	return nil
}

func EnsureAuthzDefaults() error {
	known := authz.KnownPermissions()
	if len(known) == 0 {
		return nil
	}

	return GormDB.Transaction(func(tx *gorm.DB) error {
		permissionIDs := make(map[string]string, len(known))
		for key, description := range known {
			var perm Permission
			err := tx.First(&perm, "key = ?", key).Error
			switch {
			case err == nil:
				permissionIDs[key] = perm.ID
			case errors.Is(err, gorm.ErrRecordNotFound):
				perm = Permission{ID: uuid.NewString(), Key: key, Description: description, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
				if err := tx.Create(&perm).Error; err != nil {
					return err
				}
				permissionIDs[key] = perm.ID
			default:
				return err
			}
		}

		adminRoleID, err := ensureRole(tx, authz.RoleAdmin, "Full access")
		if err != nil {
			return err
		}
		_, err = ensureRole(tx, authz.RolePlayer, "Default player role")
		if err != nil {
			return err
		}

		for _, permID := range permissionIDs {
			rp := RolePermission{
				RoleID:       adminRoleID,
				PermissionID: permID,
				CanReadSelf:  true,
				CanReadAny:   true,
				CanWriteSelf: true,
				CanWriteAny:  true,
				UpdatedAt:    time.Now().UTC(),
			}
			if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "role_id"}, {Name: "permission_id"}}, DoUpdates: clause.AssignmentColumns([]string{"can_read_self", "can_read_any", "can_write_self", "can_write_any", "updated_at"})}).Create(&rp).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func ensureRole(tx *gorm.DB, name string, description string) (string, error) {
	var role Role
	err := tx.First(&role, "name = ?", name).Error
	switch {
	case err == nil:
		if role.Description != description {
			_ = tx.Model(&Role{}).Where("id = ?", role.ID).Update("description", description).Error
		}
		return role.ID, nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		role = Role{ID: uuid.NewString(), Name: name, Description: description, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
		if err := tx.Create(&role).Error; err != nil {
			return "", err
		}
		return role.ID, nil
	default:
		return "", err
	}
}

func AssignRoleByName(accountID string, roleName string) error {
	if accountID == "" || roleName == "" {
		return nil
	}
	return GormDB.Transaction(func(tx *gorm.DB) error {
		var role Role
		if err := tx.First(&role, "name = ?", roleName).Error; err != nil {
			return err
		}
		link := AccountRole{AccountID: accountID, RoleID: role.ID}
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&link).Error
	})
}

type RolePolicyEntry struct {
	Key        string
	Capability authz.Capability
}

func LoadRolePolicyByName(roleName string) ([]RolePolicyEntry, error) {
	known := authz.KnownPermissions()
	keys := make([]string, 0, len(known))
	for key := range known {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var role Role
	if err := GormDB.First(&role, "name = ?", roleName).Error; err != nil {
		return nil, err
	}

	type row struct {
		Key          string
		CanReadSelf  bool
		CanReadAny   bool
		CanWriteSelf bool
		CanWriteAny  bool
	}
	var rows []row
	if err := GormDB.Table("role_permissions").
		Select("permissions.key as key, role_permissions.can_read_self, role_permissions.can_read_any, role_permissions.can_write_self, role_permissions.can_write_any").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ?", role.ID).
		Where("permissions.key IN ?", keys).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	lookup := make(map[string]authz.Capability, len(rows))
	for _, r := range rows {
		lookup[r.Key] = authz.Capability{ReadSelf: r.CanReadSelf, ReadAny: r.CanReadAny, WriteSelf: r.CanWriteSelf, WriteAny: r.CanWriteAny}
	}
	entries := make([]RolePolicyEntry, 0, len(keys))
	for _, key := range keys {
		entries = append(entries, RolePolicyEntry{Key: key, Capability: lookup[key]})
	}
	return entries, nil
}

func ReplaceRolePolicyByName(roleName string, capabilities map[string]authz.Capability, updatedBy *string) error {
	known := authz.KnownPermissions()
	if len(known) == 0 {
		return nil
	}
	return GormDB.Transaction(func(tx *gorm.DB) error {
		var role Role
		if err := tx.First(&role, "name = ?", roleName).Error; err != nil {
			return err
		}
		if err := tx.Model(&Role{}).Where("id = ?", role.ID).Updates(map[string]interface{}{"updated_by": updatedBy, "updated_at": time.Now().UTC()}).Error; err != nil {
			return err
		}
		for key := range known {
			var perm Permission
			if err := tx.First(&perm, "key = ?", key).Error; err != nil {
				return err
			}
			cap := capabilities[key]
			rp := RolePermission{
				RoleID:       role.ID,
				PermissionID: perm.ID,
				CanReadSelf:  cap.ReadSelf,
				CanReadAny:   cap.ReadAny,
				CanWriteSelf: cap.WriteSelf,
				CanWriteAny:  cap.WriteAny,
				UpdatedAt:    time.Now().UTC(),
			}
			if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "role_id"}, {Name: "permission_id"}}, DoUpdates: clause.AssignmentColumns([]string{"can_read_self", "can_read_any", "can_write_self", "can_write_any", "updated_at"})}).Create(&rp).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func LoadEffectivePermissions(accountID string) (map[string]authz.Capability, error) {
	if accountID == "" {
		return map[string]authz.Capability{}, nil
	}

	var roleIDs []string
	if err := GormDB.Model(&AccountRole{}).Where("account_id = ?", accountID).Pluck("role_id", &roleIDs).Error; err != nil {
		return nil, err
	}
	if len(roleIDs) == 0 {
		return map[string]authz.Capability{}, nil
	}

	type row struct {
		Key          string
		CanReadSelf  bool
		CanReadAny   bool
		CanWriteSelf bool
		CanWriteAny  bool
	}
	var rows []row
	if err := GormDB.Table("role_permissions").
		Select("permissions.key as key, role_permissions.can_read_self, role_permissions.can_read_any, role_permissions.can_write_self, role_permissions.can_write_any").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id IN ?", roleIDs).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[string]authz.Capability, len(rows))
	for _, r := range rows {
		current := result[r.Key]
		result[r.Key] = authz.MergeCapabilities(current, authz.Capability{ReadSelf: r.CanReadSelf, ReadAny: r.CanReadAny, WriteSelf: r.CanWriteSelf, WriteAny: r.CanWriteAny})
	}

	// Apply per-account overrides.
	type overrideRow struct {
		Key          string
		Mode         string
		CanReadSelf  bool
		CanReadAny   bool
		CanWriteSelf bool
		CanWriteAny  bool
	}
	var overrides []overrideRow
	if err := GormDB.Table("account_permission_overrides").
		Select("permissions.key as key, account_permission_overrides.mode, account_permission_overrides.can_read_self, account_permission_overrides.can_read_any, account_permission_overrides.can_write_self, account_permission_overrides.can_write_any").
		Joins("JOIN permissions ON permissions.id = account_permission_overrides.permission_id").
		Where("account_permission_overrides.account_id = ?", accountID).
		Scan(&overrides).Error; err != nil {
		return nil, err
	}
	for _, ov := range overrides {
		cap := result[ov.Key]
		mask := authz.Capability{ReadSelf: ov.CanReadSelf, ReadAny: ov.CanReadAny, WriteSelf: ov.CanWriteSelf, WriteAny: ov.CanWriteAny}
		switch ov.Mode {
		case PermissionOverrideAllow:
			cap = authz.MergeCapabilities(cap, mask)
		case PermissionOverrideDeny:
			if mask.ReadSelf {
				cap.ReadSelf = false
			}
			if mask.ReadAny {
				cap.ReadAny = false
			}
			if mask.WriteSelf {
				cap.WriteSelf = false
			}
			if mask.WriteAny {
				cap.WriteAny = false
			}
		}
		result[ov.Key] = cap
	}

	return result, nil
}
