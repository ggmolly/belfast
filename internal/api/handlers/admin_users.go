package handlers

import (
	"crypto/rand"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/middleware"
	"github.com/ggmolly/belfast/internal/api/response"
	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/auth"
	"github.com/ggmolly/belfast/internal/authz"
	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
)

type AdminUserHandler struct {
	Manager *auth.Manager
}

func NewAdminUserHandler(manager *auth.Manager) *AdminUserHandler {
	if manager == nil {
		manager = &auth.Manager{Config: auth.NormalizeConfig(config.AuthConfig{}), Limiter: auth.NewRateLimiter()}
	}
	manager.Config = auth.NormalizeConfig(manager.Config)
	return &AdminUserHandler{Manager: manager}
}

func RegisterAdminUserRoutes(party iris.Party, handler *AdminUserHandler) {
	party.Get("", handler.List)
	party.Get("/{id}", handler.Get)
	party.Post("", handler.Create)
	party.Patch("/{id}", handler.Update)
	party.Put("/{id}/password", handler.UpdatePassword)
	party.Delete("/{id}", handler.Delete)
}

// ListAdminUsers godoc
// @Summary     List admin users
// @Tags        Admin
// @Produce     json
// @Param       offset  query  int  false  "Pagination offset"
// @Param       limit   query  int  false  "Pagination limit"
// @Success     200  {object}  AdminUserListResponseDoc
// @Router      /api/v1/admin/users [get]
func (handler *AdminUserHandler) List(ctx iris.Context) {
	offset, _ := ctx.URLParamInt("offset")
	limit, _ := ctx.URLParamInt("limit")
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	users, total, err := orm.ListAdminAccounts(offset, limit)
	if err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to list users")
		return
	}
	responseUsers := make([]types.AdminUser, 0, len(users))
	for _, user := range users {
		responseUsers = append(responseUsers, adminUserResponse(user))
	}
	payload := types.AdminUserListResponse{
		Users: responseUsers,
		Meta: types.PaginationMeta{
			Offset: offset,
			Limit:  limit,
			Total:  total,
		},
	}
	_ = ctx.JSON(response.Success(payload))
}

// GetAdminUser godoc
// @Summary     Get admin user
// @Tags        Admin
// @Produce     json
// @Success     200  {object}  AdminUserResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Router      /api/v1/admin/users/{id} [get]
func (handler *AdminUserHandler) Get(ctx iris.Context) {
	id := ctx.Params().Get("id")
	user, err := orm.GetAccountByID(id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(ctx, iris.StatusNotFound, "not_found", "user not found")
			return
		}
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to load user")
		return
	}
	_ = ctx.JSON(response.Success(types.AdminUserResponse{User: adminUserResponse(*user)}))
}

// CreateAdminUser godoc
// @Summary     Create admin user
// @Tags        Admin
// @Accept      json
// @Produce     json
// @Param       body  body  types.AdminUserCreateRequest  true  "Admin user"
// @Success     200  {object}  AdminUserResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     409  {object}  APIErrorResponseDoc
// @Router      /api/v1/admin/users [post]
func (handler *AdminUserHandler) Create(ctx iris.Context) {
	var req types.AdminUserCreateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		writeError(ctx, iris.StatusBadRequest, "bad_request", err.Error())
		return
	}
	username := strings.TrimSpace(req.Username)
	if username == "" {
		writeError(ctx, iris.StatusBadRequest, "auth.username_required", "username required")
		return
	}
	normalized := auth.NormalizeUsername(username)
	if usernameTaken(normalized, "") {
		writeError(ctx, iris.StatusConflict, "auth.username_taken", "username already exists")
		return
	}
	cfg := handler.Manager.Config
	passwordHash, algo, err := auth.HashPassword(req.Password, cfg)
	if err != nil {
		if errors.Is(err, auth.ErrPasswordTooShort) {
			writeError(ctx, iris.StatusBadRequest, "auth.password_too_short", "password too short")
			return
		}
		if errors.Is(err, auth.ErrPasswordTooLong) {
			writeError(ctx, iris.StatusBadRequest, "auth.password_too_long", "password too long")
			return
		}
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to hash password")
		return
	}
	handle := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, handle); err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to create user handle")
		return
	}
	now := time.Now().UTC()
	user := orm.Account{
		ID:                 uuid.NewString(),
		Username:           &username,
		UsernameNormalized: &normalized,
		PasswordHash:       passwordHash,
		PasswordAlgo:       algo,
		PasswordUpdatedAt:  now,
		IsAdmin:            true,
		WebAuthnUserHandle: handle,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := orm.CreateAccount(&user); err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to create user")
		return
	}
	if err := orm.AssignRoleByName(user.ID, authz.RoleAdmin); err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to assign role")
		return
	}
	if actor, ok := middleware.GetAccount(ctx); ok {
		auth.LogAudit("user.create", &actor.ID, &user.ID, map[string]interface{}{"username": username})
	}
	_ = ctx.JSON(response.Success(types.AdminUserResponse{User: adminUserResponse(user)}))
}

// UpdateAdminUser godoc
// @Summary     Update admin user
// @Tags        Admin
// @Accept      json
// @Produce     json
// @Param       body  body  types.AdminUserUpdateRequest  true  "Admin user update"
// @Success     200  {object}  AdminUserResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     409  {object}  APIErrorResponseDoc
// @Router      /api/v1/admin/users/{id} [patch]
func (handler *AdminUserHandler) Update(ctx iris.Context) {
	id := ctx.Params().Get("id")
	var req types.AdminUserUpdateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		writeError(ctx, iris.StatusBadRequest, "bad_request", err.Error())
		return
	}
	user, err := orm.GetAccountByID(id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(ctx, iris.StatusNotFound, "not_found", "user not found")
			return
		}
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to load user")
		return
	}
	now := time.Now().UTC()
	hasUpdates := false
	if req.Username != nil {
		username := strings.TrimSpace(*req.Username)
		if username == "" {
			writeError(ctx, iris.StatusBadRequest, "auth.username_required", "username required")
			return
		}
		normalized := auth.NormalizeUsername(username)
		if usernameTaken(normalized, user.ID) {
			writeError(ctx, iris.StatusConflict, "auth.username_taken", "username already exists")
			return
		}
		if err := orm.UpdateAccountUsername(user.ID, username, normalized, now); err != nil {
			writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to update user")
			return
		}
		hasUpdates = true
	}
	if req.Disabled != nil {
		if *req.Disabled {
			if err := ensureNotLastAdmin(user.ID); err != nil {
				writeError(ctx, iris.StatusConflict, "auth.last_admin", "cannot disable last admin")
				return
			}
			disabledAt := now
			if err := orm.UpdateAccountDisabledAt(user.ID, &disabledAt, now); err != nil {
				writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to update user")
				return
			}
		} else {
			if err := orm.UpdateAccountDisabledAt(user.ID, nil, now); err != nil {
				writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to update user")
				return
			}
		}
		hasUpdates = true
	}
	if !hasUpdates {
		writeError(ctx, iris.StatusBadRequest, "bad_request", "no updates provided")
		return
	}
	user, err = orm.GetAccountByID(user.ID)
	if err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to reload user")
		return
	}
	if actor, ok := middleware.GetAccount(ctx); ok {
		auth.LogAudit("user.update", &actor.ID, &user.ID, nil)
	}
	_ = ctx.JSON(response.Success(types.AdminUserResponse{User: adminUserResponse(*user)}))
}

// UpdateAdminUserPassword godoc
// @Summary     Reset admin password
// @Tags        Admin
// @Accept      json
// @Produce     json
// @Param       body  body  types.AdminUserPasswordUpdateRequest  true  "Password"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Router      /api/v1/admin/users/{id}/password [put]
func (handler *AdminUserHandler) UpdatePassword(ctx iris.Context) {
	id := ctx.Params().Get("id")
	user, err := orm.GetAccountByID(id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(ctx, iris.StatusNotFound, "not_found", "user not found")
			return
		}
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to load user")
		return
	}
	var req types.AdminUserPasswordUpdateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		writeError(ctx, iris.StatusBadRequest, "bad_request", err.Error())
		return
	}
	cfg := handler.Manager.Config
	passwordHash, algo, err := auth.HashPassword(req.Password, cfg)
	if err != nil {
		if errors.Is(err, auth.ErrPasswordTooShort) {
			writeError(ctx, iris.StatusBadRequest, "auth.password_too_short", "password too short")
			return
		}
		if errors.Is(err, auth.ErrPasswordTooLong) {
			writeError(ctx, iris.StatusBadRequest, "auth.password_too_long", "password too long")
			return
		}
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to hash password")
		return
	}
	now := time.Now().UTC()
	if err := orm.UpdateAccountPassword(user.ID, passwordHash, algo, now, now); err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to update password")
		return
	}
	_ = auth.RevokeSessions(user.ID, "")
	if actor, ok := middleware.GetAccount(ctx); ok {
		auth.LogAudit("password.reset", &actor.ID, &user.ID, nil)
	}
	_ = ctx.JSON(response.Success(nil))
}

// DeleteAdminUser godoc
// @Summary     Delete admin user
// @Tags        Admin
// @Produce     json
// @Success     200  {object}  OKResponseDoc
// @Failure     409  {object}  APIErrorResponseDoc
// @Router      /api/v1/admin/users/{id} [delete]
func (handler *AdminUserHandler) Delete(ctx iris.Context) {
	id := ctx.Params().Get("id")
	user, err := orm.GetAccountByID(id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(ctx, iris.StatusNotFound, "not_found", "user not found")
			return
		}
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to load user")
		return
	}
	if err := ensureNotLastAdmin(user.ID); err != nil {
		writeError(ctx, iris.StatusConflict, "auth.last_admin", "cannot delete last admin")
		return
	}
	if err := orm.DeleteAccountByID(user.ID); err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to delete user")
		return
	}
	_ = orm.DeleteWebAuthnCredentialsByUserID(user.ID)
	_ = auth.RevokeSessions(user.ID, "")
	if actor, ok := middleware.GetAccount(ctx); ok {
		auth.LogAudit("user.delete", &actor.ID, &user.ID, nil)
	}
	_ = ctx.JSON(response.Success(nil))
}

func ensureNotLastAdmin(excludeID string) error {
	count, err := orm.CountEnabledAccountsWithRole(authz.RoleAdmin, excludeID)
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("last admin")
	}
	return nil
}

func usernameTaken(normalized string, excludeID string) bool {
	taken, err := orm.AccountUsernameNormalizedExists(normalized, excludeID)
	if err != nil {
		return false
	}
	return taken
}
