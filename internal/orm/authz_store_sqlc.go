package orm

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ggmolly/belfast/internal/authz"
	"github.com/ggmolly/belfast/internal/db"
	dbgen "github.com/ggmolly/belfast/internal/db/gen"
)

func listRolesSQLC() ([]Role, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	roles := make([]Role, 0, len(rows))
	for _, r := range rows {
		roles = append(roles, Role{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			CreatedAt:   r.CreatedAt.Time,
			UpdatedAt:   r.UpdatedAt.Time,
			UpdatedBy:   pgTextPtr(r.UpdatedBy),
		})
	}
	return roles, nil
}

func getRoleByNameSQLC(roleName string) (*Role, error) {
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetRoleByName(ctx, roleName)
	err = mapSQLCNotFound(err)
	if err != nil {
		return nil, err
	}
	role := Role{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
		UpdatedBy:   pgTextPtr(row.UpdatedBy),
	}
	return &role, nil
}

func listPermissionsSQLC() ([]Permission, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	perms := make([]Permission, 0, len(rows))
	for _, p := range rows {
		perms = append(perms, Permission{
			ID:          p.ID,
			Key:         p.Key,
			Description: p.Description,
			CreatedAt:   p.CreatedAt.Time,
			UpdatedAt:   p.UpdatedAt.Time,
		})
	}
	return perms, nil
}

func listAccountRoleNamesSQLC(accountID string) ([]string, error) {
	ctx := context.Background()
	return db.DefaultStore.Queries.ListAccountRoleNames(ctx, accountID)
}

func replaceAccountRolesByNameSQLC(accountID string, roleNames []string) error {
	seen := map[string]struct{}{}
	unique := make([]string, 0, len(roleNames))
	for _, name := range roleNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		unique = append(unique, name)
	}

	ctx := context.Background()
	return db.DefaultStore.WithTx(ctx, func(q *dbgen.Queries) error {
		current, err := q.ListAccountRoleNames(ctx, accountID)
		if err != nil {
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
				if err := ensureNotLastRoleSQLC(ctx, q, authz.RoleAdmin, accountID); err != nil {
					return err
				}
			}
		}

		roleIDs := make([]string, 0, len(unique))
		for _, name := range unique {
			role, err := q.GetRoleByName(ctx, name)
			err = mapSQLCNotFound(err)
			if err != nil {
				return err
			}
			roleIDs = append(roleIDs, role.ID)
		}

		if err := q.DeleteAccountRolesByAccountID(ctx, accountID); err != nil {
			return err
		}
		now := time.Now().UTC()
		for _, roleID := range roleIDs {
			if err := q.CreateAccountRoleLink(ctx, dbgen.CreateAccountRoleLinkParams{AccountID: accountID, RoleID: roleID, CreatedAt: pgTimestamptz(now)}); err != nil {
				return err
			}
		}
		return nil
	})
}

func listAccountOverridesSQLC(accountID string) ([]AccountOverrideEntry, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListAccountOverrides(ctx, accountID)
	if err != nil {
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

func replaceAccountOverridesSQLC(accountID string, overrides []AccountOverrideEntry) error {
	seen := map[string]struct{}{}
	unique := make([]AccountOverrideEntry, 0, len(overrides))
	for _, entry := range overrides {
		key := strings.TrimSpace(entry.Key)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		entry.Key = key
		unique = append(unique, entry)
	}

	ctx := context.Background()
	return db.DefaultStore.WithTx(ctx, func(q *dbgen.Queries) error {
		if err := q.DeleteAccountOverridesByAccountID(ctx, accountID); err != nil {
			return err
		}
		now := time.Now().UTC()
		for _, entry := range unique {
			perm, err := q.GetPermissionByKey(ctx, entry.Key)
			err = mapSQLCNotFound(err)
			if err != nil {
				return err
			}
			mode := entry.Mode
			if mode != PermissionOverrideAllow && mode != PermissionOverrideDeny {
				mode = PermissionOverrideAllow
			}
			if err := q.CreateAccountPermissionOverride(ctx, dbgen.CreateAccountPermissionOverrideParams{
				AccountID:    accountID,
				PermissionID: perm.ID,
				Mode:         mode,
				CanReadSelf:  entry.Capability.ReadSelf,
				CanReadAny:   entry.Capability.ReadAny,
				CanWriteSelf: entry.Capability.WriteSelf,
				CanWriteAny:  entry.Capability.WriteAny,
				UpdatedAt:    pgTimestamptz(now),
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func ensureNotLastRoleSQLC(ctx context.Context, q *dbgen.Queries, roleName string, excludeAccountID string) error {
	role, err := q.GetRoleByName(ctx, roleName)
	err = mapSQLCNotFound(err)
	if err != nil {
		return err
	}
	count, err := q.CountAccountsWithRoleExceptAccount(ctx, dbgen.CountAccountsWithRoleExceptAccountParams{RoleID: role.ID, Column2: excludeAccountID})
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("last role")
	}
	return nil
}

func ensureAuthzDefaultsSQLC() error {
	known := authz.KnownPermissions()
	if len(known) == 0 {
		return nil
	}

	ctx := context.Background()
	return db.DefaultStore.WithTx(ctx, func(q *dbgen.Queries) error {
		permissionIDs := make(map[string]string, len(known))
		now := time.Now().UTC()
		for key, description := range known {
			perm, err := q.GetPermissionByKey(ctx, key)
			err = mapSQLCNotFound(err)
			switch {
			case err == nil:
				permissionIDs[key] = perm.ID
			case errors.Is(err, db.ErrNotFound):
				id := uuid.NewString()
				if err := q.CreatePermission(ctx, dbgen.CreatePermissionParams{ID: id, Key: key, Description: description, CreatedAt: pgTimestamptz(now), UpdatedAt: pgTimestamptz(now)}); err != nil {
					return err
				}
				permissionIDs[key] = id
			default:
				return err
			}
		}

		adminRoleID, err := ensureRoleSQLC(ctx, q, authz.RoleAdmin, "Full access")
		if err != nil {
			return err
		}
		if _, err := ensureRoleSQLC(ctx, q, authz.RolePlayer, "Default player role"); err != nil {
			return err
		}

		for _, permID := range permissionIDs {
			if err := q.UpsertRolePermission(ctx, dbgen.UpsertRolePermissionParams{
				RoleID:       adminRoleID,
				PermissionID: permID,
				CanReadSelf:  true,
				CanReadAny:   true,
				CanWriteSelf: true,
				CanWriteAny:  true,
				UpdatedAt:    pgTimestamptz(now),
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func ensureRoleSQLC(ctx context.Context, q *dbgen.Queries, name string, description string) (string, error) {
	role, err := q.GetRoleByName(ctx, name)
	err = mapSQLCNotFound(err)
	switch {
	case err == nil:
		if role.Description != description {
			_ = q.UpdateRoleDescription(ctx, dbgen.UpdateRoleDescriptionParams{ID: role.ID, Description: description})
		}
		return role.ID, nil
	case errors.Is(err, db.ErrNotFound):
		id := uuid.NewString()
		now := time.Now().UTC()
		if err := q.CreateRole(ctx, dbgen.CreateRoleParams{ID: id, Name: name, Description: description, CreatedAt: pgTimestamptz(now), UpdatedAt: pgTimestamptz(now), UpdatedBy: pgTextFromPtr(nil)}); err != nil {
			return "", err
		}
		return id, nil
	default:
		return "", err
	}
}

func assignRoleByNameSQLC(accountID string, roleName string) error {
	ctx := context.Background()
	return db.DefaultStore.WithTx(ctx, func(q *dbgen.Queries) error {
		role, err := q.GetRoleByName(ctx, roleName)
		err = mapSQLCNotFound(err)
		if err != nil {
			return err
		}
		link := dbgen.CreateAccountRoleLinkParams{AccountID: accountID, RoleID: role.ID, CreatedAt: pgTimestamptz(time.Now().UTC())}
		return q.CreateAccountRoleLink(ctx, link)
	})
}

func loadRolePolicyByNameSQLC(roleName string) ([]RolePolicyEntry, error) {
	known := authz.KnownPermissions()
	keys := make([]string, 0, len(known))
	for key := range known {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	ctx := context.Background()
	role, err := db.DefaultStore.Queries.GetRoleByName(ctx, roleName)
	err = mapSQLCNotFound(err)
	if err != nil {
		return nil, err
	}
	rows, err := db.DefaultStore.Queries.ListRolePolicyRows(ctx, dbgen.ListRolePolicyRowsParams{RoleID: role.ID, Column2: keys})
	if err != nil {
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

func replaceRolePolicyByNameSQLC(roleName string, capabilities map[string]authz.Capability, updatedBy *string) error {
	known := authz.KnownPermissions()
	if len(known) == 0 {
		return nil
	}
	ctx := context.Background()
	return db.DefaultStore.WithTx(ctx, func(q *dbgen.Queries) error {
		role, err := q.GetRoleByName(ctx, roleName)
		err = mapSQLCNotFound(err)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		if err := q.UpdateRoleUpdatedBy(ctx, dbgen.UpdateRoleUpdatedByParams{ID: role.ID, UpdatedBy: pgTextFromPtr(updatedBy), UpdatedAt: pgTimestamptz(now)}); err != nil {
			return err
		}
		for key := range known {
			perm, err := q.GetPermissionByKey(ctx, key)
			err = mapSQLCNotFound(err)
			if err != nil {
				return err
			}
			cap := capabilities[key]
			if err := q.UpsertRolePermission(ctx, dbgen.UpsertRolePermissionParams{
				RoleID:       role.ID,
				PermissionID: perm.ID,
				CanReadSelf:  cap.ReadSelf,
				CanReadAny:   cap.ReadAny,
				CanWriteSelf: cap.WriteSelf,
				CanWriteAny:  cap.WriteAny,
				UpdatedAt:    pgTimestamptz(now),
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func loadEffectivePermissionsSQLC(accountID string) (map[string]authz.Capability, error) {
	ctx := context.Background()
	roleIDs, err := db.DefaultStore.Queries.ListAccountRoleIDs(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if len(roleIDs) == 0 {
		return map[string]authz.Capability{}, nil
	}
	rows, err := db.DefaultStore.Queries.ListEffectivePermissionRows(ctx, roleIDs)
	if err != nil {
		return nil, err
	}
	result := make(map[string]authz.Capability, len(rows))
	for _, r := range rows {
		current := result[r.Key]
		result[r.Key] = authz.MergeCapabilities(current, authz.Capability{ReadSelf: r.CanReadSelf, ReadAny: r.CanReadAny, WriteSelf: r.CanWriteSelf, WriteAny: r.CanWriteAny})
	}

	overrides, err := db.DefaultStore.Queries.ListAccountOverrideRows(ctx, accountID)
	if err != nil {
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
