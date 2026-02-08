package handlers

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"

	"github.com/ggmolly/belfast/internal/api/middleware"
	"github.com/ggmolly/belfast/internal/api/response"
	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/auth"
	"github.com/ggmolly/belfast/internal/authz"
	"github.com/ggmolly/belfast/internal/orm"
)

type AdminAuthzHandler struct{}

func NewAdminAuthzHandler() *AdminAuthzHandler {
	return &AdminAuthzHandler{}
}

func RegisterAdminAuthzRoutes(party iris.Party, handler *AdminAuthzHandler) {
	party.Get("/roles", handler.ListRoles)
	party.Get("/permissions", handler.ListPermissions)
	party.Get("/roles/{role}", handler.GetRolePolicy)
	party.Put("/roles/{role}", handler.ReplaceRolePolicy)
	party.Get("/accounts/{id}/roles", handler.GetAccountRoles)
	party.Put("/accounts/{id}/roles", handler.ReplaceAccountRoles)
	party.Get("/accounts/{id}/overrides", handler.GetAccountOverrides)
	party.Put("/accounts/{id}/overrides", handler.ReplaceAccountOverrides)
}

// AdminAuthzRoles godoc
// @Summary     List roles
// @Tags        Admin
// @Produce     json
// @Success     200  {object}  RoleListResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/admin/authz/roles [get]
func (handler *AdminAuthzHandler) ListRoles(ctx iris.Context) {
	roles, err := orm.ListRoles()
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list roles", nil))
		return
	}
	resp := make([]types.RoleSummary, 0, len(roles))
	for _, role := range roles {
		resp = append(resp, types.RoleSummary{
			Name:        role.Name,
			Description: role.Description,
			UpdatedAt:   role.UpdatedAt.UTC().Format(time.RFC3339),
			UpdatedBy:   derefString(role.UpdatedBy),
		})
	}
	_ = ctx.JSON(response.Success(types.RoleListResponse{Roles: resp}))
}

// AdminAuthzPermissions godoc
// @Summary     List permissions
// @Tags        Admin
// @Produce     json
// @Success     200  {object}  PermissionListResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/admin/authz/permissions [get]
func (handler *AdminAuthzHandler) ListPermissions(ctx iris.Context) {
	perms, err := orm.ListPermissions()
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list permissions", nil))
		return
	}
	resp := make([]types.PermissionSummary, 0, len(perms))
	for _, perm := range perms {
		resp = append(resp, types.PermissionSummary{Key: perm.Key, Description: perm.Description})
	}
	_ = ctx.JSON(response.Success(types.PermissionListResponse{Permissions: resp}))
}

// AdminAuthzRolePolicyGet godoc
// @Summary     Get role policy
// @Tags        Admin
// @Produce     json
// @Param       role  path  string  true  "Role name"
// @Success     200  {object}  RolePolicyResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/admin/authz/roles/{role} [get]
func (handler *AdminAuthzHandler) GetRolePolicy(ctx iris.Context) {
	roleName := strings.TrimSpace(ctx.Params().Get("role"))
	if roleName == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "role required", nil))
		return
	}
	policy, err := orm.LoadRolePolicyByName(roleName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "role not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load role policy", nil))
		return
	}
	known := authz.KnownPermissions()
	availableKeys := make([]string, 0, len(known))
	for key := range known {
		availableKeys = append(availableKeys, key)
	}
	sort.Strings(availableKeys)

	entries := make([]types.PermissionPolicyEntry, 0, len(policy))
	for _, entry := range policy {
		entries = append(entries, types.PermissionPolicyEntry{Key: entry.Key, ReadSelf: entry.Capability.ReadSelf, ReadAny: entry.Capability.ReadAny, WriteSelf: entry.Capability.WriteSelf, WriteAny: entry.Capability.WriteAny})
	}
	var role orm.Role
	_ = orm.GormDB.First(&role, "name = ?", roleName).Error
	_ = ctx.JSON(response.Success(types.RolePolicyResponse{Role: roleName, Permissions: entries, AvailableKeys: availableKeys, UpdatedAt: role.UpdatedAt.UTC().Format(time.RFC3339), UpdatedBy: derefString(role.UpdatedBy)}))
}

// AdminAuthzRolePolicyReplace godoc
// @Summary     Replace role policy
// @Tags        Admin
// @Accept      json
// @Produce     json
// @Param       role  path  string  true  "Role name"
// @Param       body  body  types.RolePolicyUpdateRequest  true  "Role policy"
// @Success     200  {object}  RolePolicyResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/admin/authz/roles/{role} [put]
func (handler *AdminAuthzHandler) ReplaceRolePolicy(ctx iris.Context) {
	roleName := strings.TrimSpace(ctx.Params().Get("role"))
	if roleName == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "role required", nil))
		return
	}
	var req types.RolePolicyUpdateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	if req.Permissions == nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "permissions required", nil))
		return
	}
	known := authz.KnownPermissions()
	seen := map[string]struct{}{}
	caps := make(map[string]authz.Capability, len(req.Permissions))
	for _, entry := range req.Permissions {
		key := strings.TrimSpace(entry.Key)
		if key == "" {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "permission key required", nil))
			return
		}
		if _, ok := known[key]; !ok {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "unknown permission key", map[string]interface{}{"key": key}))
			return
		}
		if _, ok := seen[key]; ok {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "duplicate permission key", map[string]interface{}{"key": key}))
			return
		}
		seen[key] = struct{}{}
		caps[key] = authz.Capability{ReadSelf: entry.ReadSelf, ReadAny: entry.ReadAny, WriteSelf: entry.WriteSelf, WriteAny: entry.WriteAny}
	}
	var updatedBy *string
	if account, ok := middleware.GetAccount(ctx); ok {
		updatedBy = &account.ID
	}
	if err := orm.ReplaceRolePolicyByName(roleName, caps, updatedBy); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "role not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update role policy", nil))
		return
	}
	if updatedBy != nil {
		auth.LogAudit("authz.role_policy.replace", updatedBy, nil, map[string]interface{}{"role": roleName})
	}
	handler.GetRolePolicy(ctx)
}

// AdminAuthzAccountRolesGet godoc
// @Summary     Get account roles
// @Tags        Admin
// @Produce     json
// @Param       id  path  string  true  "Account id"
// @Success     200  {object}  AccountRolesResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/admin/authz/accounts/{id}/roles [get]
func (handler *AdminAuthzHandler) GetAccountRoles(ctx iris.Context) {
	accountID := ctx.Params().Get("id")
	roles, err := orm.ListAccountRoleNames(accountID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load roles", nil))
		return
	}
	_ = ctx.JSON(response.Success(types.AccountRolesResponse{AccountID: accountID, Roles: roles}))
}

// AdminAuthzAccountRolesReplace godoc
// @Summary     Replace account roles
// @Tags        Admin
// @Accept      json
// @Produce     json
// @Param       id  path  string  true  "Account id"
// @Param       body  body  types.AccountRolesUpdateRequest  true  "Roles"
// @Success     200  {object}  AccountRolesResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/admin/authz/accounts/{id}/roles [put]
func (handler *AdminAuthzHandler) ReplaceAccountRoles(ctx iris.Context) {
	accountID := ctx.Params().Get("id")
	var req types.AccountRolesUpdateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	if req.Roles == nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "roles required", nil))
		return
	}
	var updatedBy *string
	if account, ok := middleware.GetAccount(ctx); ok {
		updatedBy = &account.ID
	}
	if err := orm.ReplaceAccountRolesByName(accountID, req.Roles); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update roles", nil))
		return
	}
	if updatedBy != nil {
		auth.LogAudit("authz.account_roles.replace", updatedBy, &accountID, map[string]interface{}{"count": len(req.Roles)})
	}
	roles, err := orm.ListAccountRoleNames(accountID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load roles", nil))
		return
	}
	_ = ctx.JSON(response.Success(types.AccountRolesResponse{AccountID: accountID, Roles: roles}))
}

// AdminAuthzAccountOverridesGet godoc
// @Summary     Get account permission overrides
// @Tags        Admin
// @Produce     json
// @Param       id  path  string  true  "Account id"
// @Success     200  {object}  AccountOverridesResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/admin/authz/accounts/{id}/overrides [get]
func (handler *AdminAuthzHandler) GetAccountOverrides(ctx iris.Context) {
	accountID := ctx.Params().Get("id")
	rows, err := orm.ListAccountOverrides(accountID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load overrides", nil))
		return
	}
	overrides := make([]types.AccountOverrideEntry, 0, len(rows))
	for _, row := range rows {
		overrides = append(overrides, types.AccountOverrideEntry{Key: row.Key, Mode: row.Mode, ReadSelf: row.Capability.ReadSelf, ReadAny: row.Capability.ReadAny, WriteSelf: row.Capability.WriteSelf, WriteAny: row.Capability.WriteAny})
	}
	_ = ctx.JSON(response.Success(types.AccountOverridesResponse{AccountID: accountID, Overrides: overrides}))
}

// AdminAuthzAccountOverridesReplace godoc
// @Summary     Replace account permission overrides
// @Tags        Admin
// @Accept      json
// @Produce     json
// @Param       id  path  string  true  "Account id"
// @Param       body  body  types.AccountOverridesUpdateRequest  true  "Overrides"
// @Success     200  {object}  AccountOverridesResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/admin/authz/accounts/{id}/overrides [put]
func (handler *AdminAuthzHandler) ReplaceAccountOverrides(ctx iris.Context) {
	accountID := ctx.Params().Get("id")
	var req types.AccountOverridesUpdateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	if req.Overrides == nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "overrides required", nil))
		return
	}
	entries := make([]orm.AccountOverrideEntry, 0, len(req.Overrides))
	for _, ov := range req.Overrides {
		mode := strings.ToLower(strings.TrimSpace(ov.Mode))
		if mode == "" {
			mode = orm.PermissionOverrideAllow
		}
		entries = append(entries, orm.AccountOverrideEntry{Key: strings.TrimSpace(ov.Key), Mode: mode, Capability: authz.Capability{ReadSelf: ov.ReadSelf, ReadAny: ov.ReadAny, WriteSelf: ov.WriteSelf, WriteAny: ov.WriteAny}})
	}
	var updatedBy *string
	if account, ok := middleware.GetAccount(ctx); ok {
		updatedBy = &account.ID
	}
	if err := orm.ReplaceAccountOverrides(accountID, entries); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update overrides", nil))
		return
	}
	if updatedBy != nil {
		auth.LogAudit("authz.account_overrides.replace", updatedBy, &accountID, map[string]interface{}{"count": len(entries)})
	}
	handler.GetAccountOverrides(ctx)
}
