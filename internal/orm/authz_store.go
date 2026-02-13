package orm

import (
	"github.com/ggmolly/belfast/internal/authz"
	"github.com/ggmolly/belfast/internal/db"
)

func ListRoles() ([]Role, error) {
	return listRolesSQLC()
}

func GetRoleByName(roleName string) (*Role, error) {
	if roleName == "" {
		return nil, db.ErrNotFound
	}
	return getRoleByNameSQLC(roleName)
}

func ListPermissions() ([]Permission, error) {
	return listPermissionsSQLC()
}

func ListAccountRoleNames(accountID string) ([]string, error) {
	if accountID == "" {
		return []string{}, nil
	}
	return listAccountRoleNamesSQLC(accountID)
}

func ReplaceAccountRolesByName(accountID string, roleNames []string) error {
	if accountID == "" {
		return nil
	}
	return replaceAccountRolesByNameSQLC(accountID, roleNames)
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
	return listAccountOverridesSQLC(accountID)
}

func ReplaceAccountOverrides(accountID string, overrides []AccountOverrideEntry) error {
	if accountID == "" {
		return nil
	}
	return replaceAccountOverridesSQLC(accountID, overrides)
}

func EnsureAuthzDefaults() error {
	return ensureAuthzDefaultsSQLC()
}

func AssignRoleByName(accountID string, roleName string) error {
	if accountID == "" || roleName == "" {
		return nil
	}
	return assignRoleByNameSQLC(accountID, roleName)
}

type RolePolicyEntry struct {
	Key        string
	Capability authz.Capability
}

func LoadRolePolicyByName(roleName string) ([]RolePolicyEntry, error) {
	return loadRolePolicyByNameSQLC(roleName)
}

func ReplaceRolePolicyByName(roleName string, capabilities map[string]authz.Capability, updatedBy *string) error {
	known := authz.KnownPermissions()
	if len(known) == 0 {
		return nil
	}
	return replaceRolePolicyByNameSQLC(roleName, capabilities, updatedBy)
}

func LoadEffectivePermissions(accountID string) (map[string]authz.Capability, error) {
	if accountID == "" {
		return map[string]authz.Capability{}, nil
	}
	return loadEffectivePermissionsSQLC(accountID)
}
