package handlers

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm"

	"github.com/ggmolly/belfast/internal/api/middleware"
	"github.com/ggmolly/belfast/internal/api/response"
	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/auth"
	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/orm"
)

type UserAuthHandler struct {
	Config   config.AuthConfig
	Limiter  *auth.RateLimiter
	Validate *validator.Validate
}

func NewUserAuthHandler(cfg *config.Config) *UserAuthHandler {
	authCfg := auth.NormalizeConfig(config.AuthConfig{})
	if cfg != nil {
		authCfg = auth.NormalizeConfig(cfg.Auth)
	}
	return &UserAuthHandler{
		Config:   authCfg,
		Limiter:  auth.NewRateLimiter(),
		Validate: validator.New(validator.WithRequiredStructEnabled()),
	}
}

func RegisterUserAuthRoutes(party iris.Party, handler *UserAuthHandler) {
	party.Post("/login", handler.Login)
	party.Post("/logout", handler.Logout)
	party.Get("/session", handler.Session)
}

// UserLogin godoc
// @Summary     User login
// @Tags        UserAuth
// @Accept      json
// @Produce     json
// @Param       body  body  types.UserAuthLoginRequest  true  "Login payload"
// @Success     200  {object}  UserAuthLoginResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     401  {object}  APIErrorResponseDoc
// @Failure     429  {object}  APIErrorResponseDoc
// @Router      /api/v1/user/auth/login [post]
func (handler *UserAuthHandler) Login(ctx iris.Context) {
	var req types.UserAuthLoginRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	if err := handler.Validate.Struct(req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "validation failed", validationErrors(err)))
		return
	}
	ip := auth.NormalizeIP(ctx.RemoteAddr())
	key := strings.Join([]string{"user_login", ip, strconv.FormatUint(uint64(req.CommanderID), 10)}, ":")
	if !handler.Limiter.Allow(key, handler.Config.RateLimitLoginMax, auth.RateLimitWindow(handler.Config)) {
		ctx.StatusCode(iris.StatusTooManyRequests)
		_ = ctx.JSON(response.Error("auth.rate_limited", "too many login attempts", nil))
		return
	}
	var account orm.Account
	if err := orm.GormDB.First(&account, "commander_id = ?", req.CommanderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusUnauthorized)
			_ = ctx.JSON(response.Error("auth.invalid_credentials", "invalid credentials", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load user", nil))
		return
	}
	valid, err := auth.VerifyPassword(req.Password, account.PasswordHash)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to verify password", nil))
		return
	}
	if !valid {
		ctx.StatusCode(iris.StatusUnauthorized)
		_ = ctx.JSON(response.Error("auth.invalid_credentials", "invalid credentials", nil))
		return
	}
	if account.DisabledAt != nil {
		ctx.StatusCode(iris.StatusForbidden)
		_ = ctx.JSON(response.Error("auth.user_disabled", "user disabled", nil))
		return
	}
	now := time.Now().UTC()
	if err := orm.GormDB.Model(&account).Update("last_login_at", now).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update login time", nil))
		return
	}
	account.LastLoginAt = &now
	session, err := auth.CreateSession(account.ID, ip, ctx.GetHeader("User-Agent"), handler.Config)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create session", nil))
		return
	}
	ctx.SetCookie(auth.BuildSessionCookie(handler.Config, session))
	auth.LogUserAudit("login.success", &account.ID, account.CommanderID, nil)
	payload := types.UserAuthLoginResponse{
		User:    userAccountResponse(account),
		Session: userSessionResponse(*session),
	}
	_ = ctx.JSON(response.Success(payload))
}

// UserLogout godoc
// @Summary     User logout
// @Tags        UserAuth
// @Produce     json
// @Success     200  {object}  OKResponseDoc
// @Failure     401  {object}  APIErrorResponseDoc
// @Router      /api/v1/user/auth/logout [post]
func (handler *UserAuthHandler) Logout(ctx iris.Context) {
	if session, ok := middleware.GetSession(ctx); ok {
		_ = auth.RevokeSession(session.ID)
	}
	ctx.SetCookie(auth.ClearSessionCookie(handler.Config))
	_ = ctx.JSON(response.Success(nil))
}

// UserSession godoc
// @Summary     Get current user session
// @Tags        UserAuth
// @Produce     json
// @Success     200  {object}  UserAuthSessionResponseDoc
// @Failure     401  {object}  APIErrorResponseDoc
// @Router      /api/v1/user/auth/session [get]
func (handler *UserAuthHandler) Session(ctx iris.Context) {
	user, ok := middleware.GetAccount(ctx)
	if !ok {
		ctx.StatusCode(iris.StatusUnauthorized)
		_ = ctx.JSON(response.Error("auth.session_missing", "session required", nil))
		return
	}
	session, ok := middleware.GetSession(ctx)
	if !ok {
		ctx.StatusCode(iris.StatusUnauthorized)
		_ = ctx.JSON(response.Error("auth.session_missing", "session required", nil))
		return
	}
	if session.CSRFToken == "" || session.CSRFExpiresAt.Before(time.Now().UTC()) {
		if token, expiresAt, err := auth.RefreshCSRF(session.ID, handler.Config); err == nil {
			session.CSRFToken = token
			session.CSRFExpiresAt = expiresAt
		}
	}
	payload := types.UserAuthSessionResponse{
		User:      userAccountResponse(*user),
		Session:   userSessionResponse(*session),
		CSRFToken: session.CSRFToken,
	}
	_ = ctx.JSON(response.Success(payload))
}

func userAccountResponse(user orm.Account) types.UserAccount {
	commanderID := uint32(0)
	if user.CommanderID != nil {
		commanderID = *user.CommanderID
	}
	return types.UserAccount{
		ID:          user.ID,
		CommanderID: commanderID,
		Disabled:    user.DisabledAt != nil,
		LastLoginAt: formatOptionalTime(user.LastLoginAt),
		CreatedAt:   user.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func userSessionResponse(session orm.Session) types.UserSession {
	return types.UserSession{
		ID:        session.ID,
		ExpiresAt: session.ExpiresAt.UTC().Format(time.RFC3339),
	}
}
