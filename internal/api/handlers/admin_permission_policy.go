package handlers

import (
	"time"

	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/middleware"
	"github.com/ggmolly/belfast/internal/api/response"
	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/auth"
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
// @Summary     Get user permission policy
// @Tags        Admin
// @Produce     json
// @Success     200  {object}  UserPermissionPolicyResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/admin/permission-policy [get]
func (handler *AdminPermissionPolicyHandler) Get(ctx iris.Context) {
	policy, err := orm.LoadUserPermissionPolicy()
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load permission policy", nil))
		return
	}
	actions, err := orm.DecodeUserPermissionActions(policy)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to parse permission policy", nil))
		return
	}
	payload := types.UserPermissionPolicyResponse{
		Actions:          actions,
		AvailableActions: append([]string{}, userPermissionActions...),
		UpdatedAt:        policy.UpdatedAt.UTC().Format(time.RFC3339),
		UpdatedBy:        derefString(policy.UpdatedBy),
	}
	_ = ctx.JSON(response.Success(payload))
}

// AdminPermissionPolicyUpdate godoc
// @Summary     Update user permission policy
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
	if req.Actions == nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "actions required", nil))
		return
	}
	actions, err := normalizeUserPermissionActions(req.Actions)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	var updatedBy *string
	if admin, ok := middleware.GetAdminUser(ctx); ok {
		updatedBy = &admin.ID
	}
	policy, err := orm.UpdateUserPermissionPolicy(actions, updatedBy)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update permission policy", nil))
		return
	}
	if updatedBy != nil {
		auth.LogAudit("user_permissions.update", updatedBy, nil, map[string]interface{}{"count": len(actions)})
	}
	payload := types.UserPermissionPolicyResponse{
		Actions:          actions,
		AvailableActions: append([]string{}, userPermissionActions...),
		UpdatedAt:        policy.UpdatedAt.UTC().Format(time.RFC3339),
		UpdatedBy:        derefString(policy.UpdatedBy),
	}
	_ = ctx.JSON(response.Success(payload))
}
