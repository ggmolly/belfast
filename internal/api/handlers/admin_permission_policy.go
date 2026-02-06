package handlers

import (
	"sort"
	"time"

	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/middleware"
	"github.com/ggmolly/belfast/internal/api/response"
	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/auth"
	"github.com/ggmolly/belfast/internal/authz"
	"github.com/ggmolly/belfast/internal/orm"
)

type AdminPermissionPolicyHandler struct{}

func NewAdminPermissionPolicyHandler() *AdminPermissionPolicyHandler {
	return &AdminPermissionPolicyHandler{}
}

func RegisterAdminPermissionPolicyRoutes(party iris.Party, handler *AdminPermissionPolicyHandler) {
	party.Get("", handler.Get)
	party.Patch("", handler.Update)
}

// AdminPermissionPolicyGet godoc
// @Summary     Get default permission policy
// @Tags        Admin
// @Produce     json
// @Success     200  {object}  UserPermissionPolicyResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/admin/permission-policy [get]
func (handler *AdminPermissionPolicyHandler) Get(ctx iris.Context) {
	policy, err := orm.LoadRolePolicyByName(authz.RolePlayer)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load permission policy", nil))
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
		entries = append(entries, types.PermissionPolicyEntry{
			Key:       entry.Key,
			ReadSelf:  entry.Capability.ReadSelf,
			ReadAny:   entry.Capability.ReadAny,
			WriteSelf: entry.Capability.WriteSelf,
			WriteAny:  entry.Capability.WriteAny,
		})
	}

	var role orm.Role
	_ = orm.GormDB.First(&role, "name = ?", authz.RolePlayer).Error
	payload := types.UserPermissionPolicyResponse{
		Role:          authz.RolePlayer,
		Permissions:   entries,
		AvailableKeys: availableKeys,
		UpdatedAt:     role.UpdatedAt.UTC().Format(time.RFC3339),
		UpdatedBy:     derefString(role.UpdatedBy),
	}
	_ = ctx.JSON(response.Success(payload))
}

// AdminPermissionPolicyUpdate godoc
// @Summary     Update default permission policy
// @Tags        Admin
// @Accept      json
// @Produce     json
// @Param       body  body  types.UserPermissionPolicyUpdateRequest  true  "Policy update"
// @Success     200  {object}  UserPermissionPolicyResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/admin/permission-policy [patch]
func (handler *AdminPermissionPolicyHandler) Update(ctx iris.Context) {
	var req types.UserPermissionPolicyUpdateRequest
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
		if entry.Key == "" {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "permission key required", nil))
			return
		}
		if _, ok := known[entry.Key]; !ok {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "unknown permission key", map[string]interface{}{"key": entry.Key}))
			return
		}
		if _, ok := seen[entry.Key]; ok {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "duplicate permission key", map[string]interface{}{"key": entry.Key}))
			return
		}
		seen[entry.Key] = struct{}{}
		caps[entry.Key] = authz.Capability{
			ReadSelf:  entry.ReadSelf,
			ReadAny:   entry.ReadAny,
			WriteSelf: entry.WriteSelf,
			WriteAny:  entry.WriteAny,
		}
	}
	var updatedBy *string
	if account, ok := middleware.GetAccount(ctx); ok {
		updatedBy = &account.ID
	}
	if err := orm.ReplaceRolePolicyByName(authz.RolePlayer, caps, updatedBy); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update permission policy", nil))
		return
	}
	if updatedBy != nil {
		auth.LogAudit("permissions.update", updatedBy, nil, map[string]interface{}{"role": authz.RolePlayer})
	}
	rolePolicy, err := orm.LoadRolePolicyByName(authz.RolePlayer)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load permission policy", nil))
		return
	}
	entries := make([]types.PermissionPolicyEntry, 0, len(rolePolicy))
	for _, entry := range rolePolicy {
		entries = append(entries, types.PermissionPolicyEntry{
			Key:       entry.Key,
			ReadSelf:  entry.Capability.ReadSelf,
			ReadAny:   entry.Capability.ReadAny,
			WriteSelf: entry.Capability.WriteSelf,
			WriteAny:  entry.Capability.WriteAny,
		})
	}
	availableKeys := make([]string, 0, len(known))
	for key := range known {
		availableKeys = append(availableKeys, key)
	}
	sort.Strings(availableKeys)
	var role orm.Role
	_ = orm.GormDB.First(&role, "name = ?", authz.RolePlayer).Error
	payload := types.UserPermissionPolicyResponse{
		Role:          authz.RolePlayer,
		Permissions:   entries,
		AvailableKeys: availableKeys,
		UpdatedAt:     role.UpdatedAt.UTC().Format(time.RFC3339),
		UpdatedBy:     derefString(role.UpdatedBy),
	}
	_ = ctx.JSON(response.Success(payload))
}
