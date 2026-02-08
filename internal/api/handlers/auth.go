package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/ggmolly/belfast/internal/api/middleware"
	"github.com/ggmolly/belfast/internal/api/response"
	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/auth"
	"github.com/ggmolly/belfast/internal/authz"
	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/orm"
)

type AuthHandler struct {
	Manager *auth.Manager
}

func NewAuthHandler(manager *auth.Manager) *AuthHandler {
	if manager == nil {
		manager = &auth.Manager{Config: auth.NormalizeConfig(config.AuthConfig{}), Limiter: auth.NewRateLimiter(), Selection: protocol.AuthenticatorSelection{UserVerification: protocol.VerificationPreferred}}
	}
	if manager.Limiter == nil {
		manager.Limiter = auth.NewRateLimiter()
	}
	manager.Config = auth.NormalizeConfig(manager.Config)
	return &AuthHandler{Manager: manager}
}

func RegisterAuthRoutes(party iris.Party, handler *AuthHandler) {
	party.Post("/bootstrap", handler.Bootstrap)
	party.Get("/bootstrap/status", handler.BootstrapStatus)
	party.Post("/login", handler.Login)
	party.Post("/logout", handler.Logout)
	party.Get("/session", handler.Session)
	party.Post("/password/change", handler.ChangePassword)
	party.Post("/passkeys/register/options", handler.PasskeyRegisterOptions)
	party.Post("/passkeys/register/verify", handler.PasskeyRegisterVerify)
	party.Post("/passkeys/authenticate/options", handler.PasskeyAuthenticateOptions)
	party.Post("/passkeys/authenticate/verify", handler.PasskeyAuthenticateVerify)
	party.Get("/passkeys", handler.PasskeyList)
	party.Delete("/passkeys/{credential_id}", handler.PasskeyDelete)
}

// BootstrapStatus godoc
// @Summary     Check bootstrap status
// @Tags        Auth
// @Produce     json
// @Success     200  {object}  AuthBootstrapStatusResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/auth/bootstrap/status [get]
func (handler *AuthHandler) BootstrapStatus(ctx iris.Context) {
	var count int64
	if err := orm.GormDB.Model(&orm.Account{}).Where("is_admin = ?", true).Count(&count).Error; err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to check admin users")
		return
	}
	payload := types.AuthBootstrapStatusResponse{
		CanBootstrap: count == 0,
		AdminCount:   count,
	}
	_ = ctx.JSON(response.Success(payload))
}

// Bootstrap godoc
// @Summary     Create first admin
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       body  body  types.AuthBootstrapRequest  true  "Bootstrap payload"
// @Success     200   {object}  AuthLoginResponseDoc
// @Failure     400   {object}  APIErrorResponseDoc
// @Failure     409   {object}  APIErrorResponseDoc
// @Router      /api/v1/auth/bootstrap [post]
func (handler *AuthHandler) Bootstrap(ctx iris.Context) {
	var req types.AuthBootstrapRequest
	if err := ctx.ReadJSON(&req); err != nil {
		writeError(ctx, iris.StatusBadRequest, "bad_request", err.Error())
		return
	}
	username := strings.TrimSpace(req.Username)
	if username == "" {
		writeError(ctx, iris.StatusBadRequest, "auth.username_required", "username required")
		return
	}
	var count int64
	if err := orm.GormDB.Model(&orm.Account{}).Where("is_admin = ?", true).Count(&count).Error; err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to check admin users")
		return
	}
	if count > 0 {
		writeError(ctx, iris.StatusConflict, "auth.bootstrap_closed", "bootstrap already completed")
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
	normalized := auth.NormalizeUsername(username)
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
	if err := orm.GormDB.Create(&user).Error; err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to create admin user")
		return
	}
	if err := orm.AssignRoleByName(user.ID, authz.RoleAdmin); err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to assign admin role")
		return
	}
	session, err := auth.CreateSession(user.ID, auth.NormalizeIP(ctx.RemoteAddr()), ctx.GetHeader("User-Agent"), cfg)
	if err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to create session")
		return
	}
	ctx.SetCookie(auth.BuildSessionCookie(cfg, session))
	auth.LogAudit("bootstrap", &user.ID, &user.ID, map[string]interface{}{"username": derefString(user.Username)})
	payload := types.AuthLoginResponse{
		User:    adminUserResponse(user),
		Session: authSessionResponse(*session),
	}
	_ = ctx.JSON(response.Success(payload))
}

// Login godoc
// @Summary     Admin login
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       body  body  types.AuthLoginRequest  true  "Login payload"
// @Success     200   {object}  AuthLoginResponseDoc
// @Failure     401   {object}  APIErrorResponseDoc
// @Failure     429   {object}  APIErrorResponseDoc
// @Router      /api/v1/auth/login [post]
func (handler *AuthHandler) Login(ctx iris.Context) {
	var req types.AuthLoginRequest
	if err := ctx.ReadJSON(&req); err != nil {
		writeError(ctx, iris.StatusBadRequest, "bad_request", err.Error())
		return
	}
	username := strings.TrimSpace(req.Username)
	if username == "" {
		writeError(ctx, iris.StatusBadRequest, "auth.username_required", "username required")
		return
	}
	cfg := handler.Manager.Config
	key := auth.NormalizeIP(ctx.RemoteAddr()) + ":" + auth.NormalizeUsername(username)
	if !handler.Manager.Limiter.Allow(key, cfg.RateLimitLoginMax, auth.RateLimitWindow(cfg)) {
		writeError(ctx, iris.StatusTooManyRequests, "auth.rate_limited", "too many login attempts")
		return
	}
	var user orm.Account
	if err := orm.GormDB.First(&user, "username_normalized = ?", auth.NormalizeUsername(username)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(ctx, iris.StatusUnauthorized, "auth.invalid_credentials", "invalid credentials")
			return
		}
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to load admin user")
		return
	}
	valid, err := auth.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to verify password")
		return
	}
	if !valid {
		auth.LogAudit("login.fail", nil, nil, map[string]interface{}{"username": username})
		writeError(ctx, iris.StatusUnauthorized, "auth.invalid_credentials", "invalid credentials")
		return
	}
	if user.DisabledAt != nil {
		writeError(ctx, iris.StatusForbidden, "auth.user_disabled", "user disabled")
		return
	}
	now := time.Now().UTC()
	if err := orm.GormDB.Model(&user).Update("last_login_at", now).Error; err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to update login time")
		return
	}
	user.LastLoginAt = &now
	session, err := auth.CreateSession(user.ID, auth.NormalizeIP(ctx.RemoteAddr()), ctx.GetHeader("User-Agent"), cfg)
	if err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to create session")
		return
	}
	ctx.SetCookie(auth.BuildSessionCookie(cfg, session))
	auth.LogAudit("login.success", &user.ID, &user.ID, map[string]interface{}{"username": derefString(user.Username)})
	payload := types.AuthLoginResponse{
		User:    adminUserResponse(user),
		Session: authSessionResponse(*session),
	}
	_ = ctx.JSON(response.Success(payload))
}

// Logout godoc
// @Summary     Logout admin
// @Tags        Auth
// @Produce     json
// @Success     200  {object}  OKResponseDoc
// @Failure     401  {object}  APIErrorResponseDoc
// @Router      /api/v1/auth/logout [post]
func (handler *AuthHandler) Logout(ctx iris.Context) {
	if session, ok := middleware.GetSession(ctx); ok {
		_ = auth.RevokeSession(session.ID)
	}
	ctx.SetCookie(auth.ClearSessionCookie(handler.Manager.Config))
	_ = ctx.JSON(response.Success(nil))
}

// Session godoc
// @Summary     Get current session
// @Tags        Auth
// @Produce     json
// @Success     200  {object}  AuthSessionResponseDoc
// @Failure     401  {object}  APIErrorResponseDoc
// @Router      /api/v1/auth/session [get]
func (handler *AuthHandler) Session(ctx iris.Context) {
	user, ok := middleware.GetAccount(ctx)
	if !ok {
		writeError(ctx, iris.StatusUnauthorized, "auth.session_missing", "session required")
		return
	}
	session, ok := middleware.GetSession(ctx)
	if !ok {
		writeError(ctx, iris.StatusUnauthorized, "auth.session_missing", "session required")
		return
	}
	if session.CSRFToken == "" || session.CSRFExpiresAt.Before(time.Now().UTC()) {
		if token, expiresAt, err := auth.RefreshCSRF(session.ID, handler.Manager.Config); err == nil {
			session.CSRFToken = token
			session.CSRFExpiresAt = expiresAt
		}
	}
	payload := types.AuthSessionResponse{
		User:      adminUserResponse(*user),
		Session:   authSessionResponse(*session),
		CSRFToken: session.CSRFToken,
	}
	_ = ctx.JSON(response.Success(payload))
}

// ChangePassword godoc
// @Summary     Change password
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       body  body  types.AuthPasswordChangeRequest  true  "Password change"
// @Success     200  {object}  PasskeyListResponseDoc
// @Failure     401  {object}  APIErrorResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Router      /api/v1/auth/password/change [post]
func (handler *AuthHandler) ChangePassword(ctx iris.Context) {
	user, ok := middleware.GetAccount(ctx)
	if !ok {
		writeError(ctx, iris.StatusUnauthorized, "auth.session_missing", "session required")
		return
	}
	session, ok := middleware.GetSession(ctx)
	if !ok {
		writeError(ctx, iris.StatusUnauthorized, "auth.session_missing", "session required")
		return
	}
	var req types.AuthPasswordChangeRequest
	if err := ctx.ReadJSON(&req); err != nil {
		writeError(ctx, iris.StatusBadRequest, "bad_request", err.Error())
		return
	}
	valid, err := auth.VerifyPassword(req.CurrentPassword, user.PasswordHash)
	if err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to verify password")
		return
	}
	if !valid {
		writeError(ctx, iris.StatusUnauthorized, "auth.invalid_credentials", "invalid credentials")
		return
	}
	cfg := handler.Manager.Config
	passwordHash, algo, err := auth.HashPassword(req.NewPassword, cfg)
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
	updates := map[string]interface{}{
		"password_hash":       passwordHash,
		"password_algo":       algo,
		"password_updated_at": time.Now().UTC(),
	}
	if err := orm.GormDB.Model(&orm.Account{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to update password")
		return
	}
	_ = auth.RevokeSessions(user.ID, session.ID)
	auth.LogAudit("password.change", &user.ID, &user.ID, nil)
	_ = ctx.JSON(response.Success(nil))
}

// PasskeyRegisterOptions godoc
// @Summary     Begin passkey registration
// @Tags        Passkeys
// @Accept      json
// @Produce     json
// @Param       body  body  types.PasskeyRegisterOptionsRequest  true  "Passkey options"
// @Success     200  {object}  PasskeyRegisterOptionsResponseDoc
// @Failure     401  {object}  APIErrorResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Router      /api/v1/auth/passkeys/register/options [post]
func (handler *AuthHandler) PasskeyRegisterOptions(ctx iris.Context) {
	if err := handler.Manager.EnsureWebAuthn(); err != nil {
		writeError(ctx, iris.StatusInternalServerError, "auth.webauthn_not_configured", "webauthn not configured")
		return
	}
	user, ok := middleware.GetAccount(ctx)
	if !ok {
		writeError(ctx, iris.StatusUnauthorized, "auth.session_missing", "session required")
		return
	}
	var req types.PasskeyRegisterOptionsRequest
	if err := ctx.ReadJSON(&req); err != nil {
		writeError(ctx, iris.StatusBadRequest, "bad_request", err.Error())
		return
	}
	if err := auth.EnsureUserHandle(user); err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to ensure user handle")
		return
	}
	credentials, err := loadUserCredentials(user.ID)
	if err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to load credentials")
		return
	}
	webUser, err := auth.BuildWebAuthnUser(*user, credentials)
	if err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to prepare user")
		return
	}
	options := make([]webauthn.RegistrationOption, 0, 3)
	if len(credentials) > 0 {
		exclusions, err := buildCredentialDescriptors(credentials)
		if err != nil {
			writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to build exclusions")
			return
		}
		options = append(options, webauthn.WithExclusions(exclusions))
	}
	selection := handler.Manager.Selection
	if req.UserVerification != nil {
		selection.UserVerification = protocol.UserVerificationRequirement(*req.UserVerification)
	}
	if req.ResidentKey != nil {
		selection.ResidentKey = protocol.ResidentKeyRequirement(*req.ResidentKey)
		if selection.ResidentKey == protocol.ResidentKeyRequirementRequired {
			selection.RequireResidentKey = protocol.ResidentKeyRequired()
		} else {
			selection.RequireResidentKey = protocol.ResidentKeyNotRequired()
		}
	}
	options = append(options, webauthn.WithAuthenticatorSelection(selection))
	creation, sessionData, err := handler.Manager.WebAuthn.BeginRegistration(webUser, options...)
	if err != nil {
		writeError(ctx, iris.StatusBadRequest, "auth.webauthn_verification_failed", err.Error())
		return
	}
	challengeExpires := time.Now().UTC().Add(auth.WebAuthnChallengeTTL(handler.Manager.Config))
	sessionData.Expires = challengeExpires
	if _, err := auth.StoreChallenge(&user.ID, "webauthn.register", *sessionData, challengeExpires); err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to store challenge")
		return
	}
	_ = ctx.JSON(response.Success(creation))
}

// PasskeyRegisterVerify godoc
// @Summary     Finish passkey registration
// @Tags        Passkeys
// @Accept      json
// @Produce     json
// @Param       body  body  types.PasskeyRegisterVerifyRequest  true  "Passkey registration"
// @Success     200  {object}  PasskeyRegisterResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     401  {object}  APIErrorResponseDoc
// @Router      /api/v1/auth/passkeys/register/verify [post]
func (handler *AuthHandler) PasskeyRegisterVerify(ctx iris.Context) {
	if err := handler.Manager.EnsureWebAuthn(); err != nil {
		writeError(ctx, iris.StatusInternalServerError, "auth.webauthn_not_configured", "webauthn not configured")
		return
	}
	user, ok := middleware.GetAccount(ctx)
	if !ok {
		writeError(ctx, iris.StatusUnauthorized, "auth.session_missing", "session required")
		return
	}
	cfg := handler.Manager.Config
	key := auth.NormalizeIP(ctx.RemoteAddr()) + ":" + user.ID
	if !handler.Manager.Limiter.Allow(key, cfg.RateLimitPasskeyMax, auth.RateLimitWindow(cfg)) {
		writeError(ctx, iris.StatusTooManyRequests, "auth.rate_limited", "too many attempts")
		return
	}
	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		writeError(ctx, iris.StatusBadRequest, "bad_request", "failed to read request")
		return
	}
	var req types.PasskeyRegisterVerifyRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(ctx, iris.StatusBadRequest, "bad_request", err.Error())
		return
	}
	challengeValue, err := auth.ExtractChallenge(req.Credential.Response.ClientDataJSON)
	if err != nil {
		writeError(ctx, iris.StatusBadRequest, "auth.webauthn_verification_failed", "invalid client data")
		return
	}
	challenge, sessionData, err := auth.LoadChallengeByUser(user.ID, "webauthn.register")
	if err != nil {
		writeError(ctx, iris.StatusBadRequest, "auth.challenge_expired", "challenge expired")
		return
	}
	if challenge.ExpiresAt.Before(time.Now().UTC()) {
		_ = auth.DeleteChallenge(challenge.ID)
		writeError(ctx, iris.StatusBadRequest, "auth.challenge_expired", "challenge expired")
		return
	}
	if challengeValue != challenge.Challenge {
		writeError(ctx, iris.StatusBadRequest, "auth.webauthn_verification_failed", "challenge mismatch")
		return
	}
	if existsCredential(req.Credential.ID) {
		writeError(ctx, iris.StatusConflict, "auth.credential_exists", "credential already exists")
		return
	}
	credentialBody, err := json.Marshal(req.Credential)
	if err != nil {
		writeError(ctx, iris.StatusBadRequest, "bad_request", "failed to read credential")
		return
	}
	ctx.Request().Body = io.NopCloser(bytes.NewReader(credentialBody))
	credentials, err := loadUserCredentials(user.ID)
	if err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to load credentials")
		return
	}
	webUser, err := auth.BuildWebAuthnUser(*user, credentials)
	if err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to prepare user")
		return
	}
	credential, err := handler.Manager.WebAuthn.FinishRegistration(webUser, *sessionData, ctx.Request())
	if err != nil {
		writeError(ctx, iris.StatusBadRequest, "auth.webauthn_verification_failed", "registration failed")
		return
	}
	createdAt := time.Now().UTC()
	credentialID := base64.RawURLEncoding.EncodeToString(credential.ID)
	transports := make([]string, 0, len(credential.Transport))
	for _, transport := range credential.Transport {
		transports = append(transports, string(transport))
	}
	var aaguid string
	if len(credential.Authenticator.AAGUID) > 0 {
		aaguid = base64.RawURLEncoding.EncodeToString(credential.Authenticator.AAGUID)
	}
	backupEligible := credential.Flags.BackupEligible
	backupState := credential.Flags.BackupState
	record := orm.WebAuthnCredential{
		ID:             uuid.NewString(),
		UserID:         user.ID,
		CredentialID:   credentialID,
		PublicKey:      credential.PublicKey,
		SignCount:      credential.Authenticator.SignCount,
		Transports:     transports,
		AAGUID:         aaguid,
		AttestationFmt: credential.AttestationType,
		ResidentKey:    false,
		BackupEligible: &backupEligible,
		BackupState:    &backupState,
		CreatedAt:      createdAt,
		Label:          req.Label,
		RPID:           handler.Manager.Config.WebAuthnRPID,
	}
	if err := orm.GormDB.Create(&record).Error; err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to store credential")
		return
	}
	_ = auth.DeleteChallenge(challenge.ID)
	auth.LogAudit("passkey.add", &user.ID, &user.ID, map[string]interface{}{"credential_id": credentialID})
	payload := types.PasskeyRegisterResponse{
		CredentialID: credentialID,
		Label:        derefString(req.Label),
		CreatedAt:    createdAt.Format(time.RFC3339),
	}
	_ = ctx.JSON(response.Success(payload))
}

// PasskeyAuthenticateOptions godoc
// @Summary     Begin passkey authentication
// @Tags        Passkeys
// @Accept      json
// @Produce     json
// @Param       body  body  types.PasskeyAuthenticateOptionsRequest  true  "Passkey auth options"
// @Success     200  {object}  PasskeyAuthenticateOptionsResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Router      /api/v1/auth/passkeys/authenticate/options [post]
func (handler *AuthHandler) PasskeyAuthenticateOptions(ctx iris.Context) {
	if err := handler.Manager.EnsureWebAuthn(); err != nil {
		writeError(ctx, iris.StatusInternalServerError, "auth.webauthn_not_configured", "webauthn not configured")
		return
	}
	var req types.PasskeyAuthenticateOptionsRequest
	if err := ctx.ReadJSON(&req); err != nil {
		writeError(ctx, iris.StatusBadRequest, "bad_request", err.Error())
		return
	}
	var (
		assertion   *protocol.CredentialAssertion
		sessionData *webauthn.SessionData
		err         error
		userID      *string
	)
	if req.Username != nil {
		username := strings.TrimSpace(*req.Username)
		if username == "" {
			writeError(ctx, iris.StatusBadRequest, "auth.user_not_found", "user not found")
			return
		}
		user, credentials, err := loadUserWithCredentials(username)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				writeError(ctx, iris.StatusNotFound, "auth.user_not_found", "user not found")
				return
			}
			writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to load user")
			return
		}
		if user.DisabledAt != nil {
			writeError(ctx, iris.StatusForbidden, "auth.user_disabled", "user disabled")
			return
		}
		if err := auth.EnsureUserHandle(user); err != nil {
			writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to ensure user handle")
			return
		}
		webUser, err := auth.BuildWebAuthnUser(*user, credentials)
		if err != nil {
			writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to prepare user")
			return
		}
		assertion, sessionData, err = handler.Manager.WebAuthn.BeginLogin(webUser)
		userID = &user.ID
	} else {
		assertion, sessionData, err = handler.Manager.WebAuthn.BeginDiscoverableLogin()
	}
	if err != nil {
		writeError(ctx, iris.StatusBadRequest, "auth.passkey_not_found", "no credentials")
		return
	}
	challengeExpires := time.Now().UTC().Add(auth.WebAuthnChallengeTTL(handler.Manager.Config))
	sessionData.Expires = challengeExpires
	if _, err := auth.StoreChallenge(userID, "webauthn.auth", *sessionData, challengeExpires); err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to store challenge")
		return
	}
	_ = ctx.JSON(response.Success(assertion))
}

// PasskeyAuthenticateVerify godoc
// @Summary     Finish passkey authentication
// @Tags        Passkeys
// @Accept      json
// @Produce     json
// @Param       body  body  types.PasskeyAuthenticateVerifyRequest  true  "Passkey auth verify"
// @Success     200  {object}  AuthLoginResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     401  {object}  APIErrorResponseDoc
// @Router      /api/v1/auth/passkeys/authenticate/verify [post]
func (handler *AuthHandler) PasskeyAuthenticateVerify(ctx iris.Context) {
	if err := handler.Manager.EnsureWebAuthn(); err != nil {
		writeError(ctx, iris.StatusInternalServerError, "auth.webauthn_not_configured", "webauthn not configured")
		return
	}
	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		writeError(ctx, iris.StatusBadRequest, "bad_request", "failed to read request")
		return
	}
	var req types.PasskeyAuthenticateVerifyRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(ctx, iris.StatusBadRequest, "bad_request", err.Error())
		return
	}
	cfg := handler.Manager.Config
	key := auth.NormalizeIP(ctx.RemoteAddr())
	if req.Username != nil {
		key += ":" + auth.NormalizeUsername(*req.Username)
	}
	if !handler.Manager.Limiter.Allow(key, cfg.RateLimitPasskeyMax, auth.RateLimitWindow(cfg)) {
		writeError(ctx, iris.StatusTooManyRequests, "auth.rate_limited", "too many attempts")
		return
	}
	challengeValue, err := auth.ExtractChallenge(req.Credential.Response.ClientDataJSON)
	if err != nil {
		writeError(ctx, iris.StatusBadRequest, "auth.webauthn_verification_failed", "invalid client data")
		return
	}
	credentialBody, err := json.Marshal(req.Credential)
	if err != nil {
		writeError(ctx, iris.StatusBadRequest, "bad_request", "failed to read credential")
		return
	}
	ctx.Request().Body = io.NopCloser(bytes.NewReader(credentialBody))
	var (
		challenge   *orm.AuthChallenge
		sessionData *webauthn.SessionData
	)
	if req.Username != nil {
		username := strings.TrimSpace(*req.Username)
		if username == "" {
			writeError(ctx, iris.StatusNotFound, "auth.user_not_found", "user not found")
			return
		}
		user, credentials, err := loadUserWithCredentials(username)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				writeError(ctx, iris.StatusNotFound, "auth.user_not_found", "user not found")
				return
			}
			writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to load user")
			return
		}
		if user.DisabledAt != nil {
			writeError(ctx, iris.StatusForbidden, "auth.user_disabled", "user disabled")
			return
		}
		if err := auth.EnsureUserHandle(user); err != nil {
			writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to ensure user handle")
			return
		}
		if !credentialExistsForUser(user.ID, req.Credential.ID) {
			writeError(ctx, iris.StatusNotFound, "auth.passkey_not_found", "passkey not found")
			return
		}
		challenge, sessionData, err = auth.LoadChallengeByUser(user.ID, "webauthn.auth")
		if err != nil {
			writeError(ctx, iris.StatusBadRequest, "auth.challenge_expired", "challenge expired")
			return
		}
		if challenge.ExpiresAt.Before(time.Now().UTC()) {
			_ = auth.DeleteChallenge(challenge.ID)
			writeError(ctx, iris.StatusBadRequest, "auth.challenge_expired", "challenge expired")
			return
		}
		if challengeValue != challenge.Challenge {
			writeError(ctx, iris.StatusBadRequest, "auth.webauthn_verification_failed", "challenge mismatch")
			return
		}
		webUser, err := auth.BuildWebAuthnUser(*user, credentials)
		if err != nil {
			writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to prepare user")
			return
		}
		credential, err := handler.Manager.WebAuthn.FinishLogin(webUser, *sessionData, ctx.Request())
		if err != nil {
			writeError(ctx, iris.StatusBadRequest, "auth.webauthn_verification_failed", "authentication failed")
			return
		}
		_ = auth.DeleteChallenge(challenge.ID)
		if err := updateCredentialUsage(credential); err != nil {
			writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to update credential")
			return
		}
		if err := orm.GormDB.Model(&orm.Account{}).Where("id = ?", user.ID).Update("last_login_at", time.Now().UTC()).Error; err != nil {
			writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to update login time")
			return
		}
		session, err := auth.CreateSession(user.ID, auth.NormalizeIP(ctx.RemoteAddr()), ctx.GetHeader("User-Agent"), handler.Manager.Config)
		if err != nil {
			writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to create session")
			return
		}
		ctx.SetCookie(auth.BuildSessionCookie(handler.Manager.Config, session))
		auth.LogAudit("login.success", &user.ID, &user.ID, map[string]interface{}{"method": "passkey"})
		payload := types.AuthLoginResponse{
			User:    adminUserResponse(*user),
			Session: authSessionResponse(*session),
		}
		_ = ctx.JSON(response.Success(payload))
		return
	}
	challenge, sessionData, err = auth.LoadChallengeByChallenge(challengeValue, "webauthn.auth")
	if err != nil {
		writeError(ctx, iris.StatusBadRequest, "auth.challenge_expired", "challenge expired")
		return
	}
	if challenge.ExpiresAt.Before(time.Now().UTC()) {
		_ = auth.DeleteChallenge(challenge.ID)
		writeError(ctx, iris.StatusBadRequest, "auth.challenge_expired", "challenge expired")
		return
	}
	var account orm.Account
	errUserDisabled := errors.New("user disabled")
	userHandler := func(rawID, userHandle []byte) (webauthn.User, error) {
		if err := orm.GormDB.First(&account, "webauthn_user_handle = ?", userHandle).Error; err != nil {
			return nil, err
		}
		if account.DisabledAt != nil {
			return nil, errUserDisabled
		}
		credentials, err := loadUserCredentials(account.ID)
		if err != nil {
			return nil, err
		}
		return auth.BuildWebAuthnUser(account, credentials)
	}
	user, credential, err := handler.Manager.WebAuthn.FinishPasskeyLogin(userHandler, *sessionData, ctx.Request())
	if err != nil {
		if errors.Is(err, errUserDisabled) {
			writeError(ctx, iris.StatusForbidden, "auth.user_disabled", "user disabled")
			return
		}
		writeError(ctx, iris.StatusBadRequest, "auth.webauthn_verification_failed", "authentication failed")
		return
	}
	if credential == nil || user == nil {
		writeError(ctx, iris.StatusBadRequest, "auth.webauthn_verification_failed", "authentication failed")
		return
	}
	_ = auth.DeleteChallenge(challenge.ID)
	if err := updateCredentialUsage(credential); err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to update credential")
		return
	}
	if account.ID == "" {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to resolve user")
		return
	}
	if err := orm.GormDB.Model(&orm.Account{}).Where("id = ?", account.ID).Update("last_login_at", time.Now().UTC()).Error; err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to update login time")
		return
	}
	session, err := auth.CreateSession(account.ID, auth.NormalizeIP(ctx.RemoteAddr()), ctx.GetHeader("User-Agent"), handler.Manager.Config)
	if err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to create session")
		return
	}
	ctx.SetCookie(auth.BuildSessionCookie(handler.Manager.Config, session))
	auth.LogAudit("login.success", &account.ID, &account.ID, map[string]interface{}{"method": "passkey"})
	payload := types.AuthLoginResponse{
		User:    adminUserResponse(account),
		Session: authSessionResponse(*session),
	}
	_ = ctx.JSON(response.Success(payload))
}

// PasskeyList godoc
// @Summary     List passkeys
// @Tags        Passkeys
// @Produce     json
// @Success     200  {object}  OKResponseDoc
// @Failure     401  {object}  APIErrorResponseDoc
// @Router      /api/v1/auth/passkeys [get]
func (handler *AuthHandler) PasskeyList(ctx iris.Context) {
	user, ok := middleware.GetAccount(ctx)
	if !ok {
		writeError(ctx, iris.StatusUnauthorized, "auth.session_missing", "session required")
		return
	}
	credentials, err := loadUserCredentials(user.ID)
	if err != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to load credentials")
		return
	}
	passkeys := make([]types.PasskeySummary, 0, len(credentials))
	for _, credential := range credentials {
		passkeys = append(passkeys, types.PasskeySummary{
			CredentialID:   credential.CredentialID,
			Label:          derefString(credential.Label),
			CreatedAt:      credential.CreatedAt.UTC().Format(time.RFC3339),
			LastUsedAt:     formatOptionalTime(credential.LastUsedAt),
			Transports:     credential.Transports,
			AAGUID:         credential.AAGUID,
			BackupEligible: credential.BackupEligible,
			BackupState:    credential.BackupState,
		})
	}
	_ = ctx.JSON(response.Success(types.PasskeyListResponse{Passkeys: passkeys}))
}

// PasskeyDelete godoc
// @Summary     Remove passkey
// @Tags        Passkeys
// @Produce     json
// @Success     200  {object}  OKResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Router      /api/v1/auth/passkeys/{credential_id} [delete]
func (handler *AuthHandler) PasskeyDelete(ctx iris.Context) {
	user, ok := middleware.GetAccount(ctx)
	if !ok {
		writeError(ctx, iris.StatusUnauthorized, "auth.session_missing", "session required")
		return
	}
	credentialID := ctx.Params().Get("credential_id")
	if credentialID == "" {
		writeError(ctx, iris.StatusBadRequest, "bad_request", "credential_id required")
		return
	}
	result := orm.GormDB.Delete(&orm.WebAuthnCredential{}, "user_id = ? AND credential_id = ?", user.ID, credentialID)
	if result.Error != nil {
		writeError(ctx, iris.StatusInternalServerError, "internal_error", "failed to delete credential")
		return
	}
	if result.RowsAffected == 0 {
		writeError(ctx, iris.StatusNotFound, "auth.passkey_not_found", "passkey not found")
		return
	}
	auth.LogAudit("passkey.remove", &user.ID, &user.ID, map[string]interface{}{"credential_id": credentialID})
	_ = ctx.JSON(response.Success(nil))
}

func writeError(ctx iris.Context, status int, code string, message string) {
	ctx.StatusCode(status)
	_ = ctx.JSON(response.Error(code, message, nil))
}

func adminUserResponse(user orm.Account) types.AdminUser {
	username := ""
	if user.Username != nil {
		username = *user.Username
	}
	return types.AdminUser{
		ID:          user.ID,
		Username:    username,
		IsAdmin:     user.IsAdmin,
		Disabled:    user.DisabledAt != nil,
		LastLoginAt: formatOptionalTime(user.LastLoginAt),
		CreatedAt:   user.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func authSessionResponse(session orm.Session) types.AuthSession {
	return types.AuthSession{
		ID:        session.ID,
		ExpiresAt: session.ExpiresAt.UTC().Format(time.RFC3339),
	}
}

func formatOptionalTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func loadUserCredentials(userID string) ([]orm.WebAuthnCredential, error) {
	var credentials []orm.WebAuthnCredential
	if err := orm.GormDB.Where("user_id = ?", userID).Find(&credentials).Error; err != nil {
		return nil, err
	}
	return credentials, nil
}

func loadUserWithCredentials(username string) (*orm.Account, []orm.WebAuthnCredential, error) {
	var user orm.Account
	if err := orm.GormDB.First(&user, "username_normalized = ?", auth.NormalizeUsername(username)).Error; err != nil {
		return nil, nil, err
	}
	credentials, err := loadUserCredentials(user.ID)
	if err != nil {
		return nil, nil, err
	}
	return &user, credentials, nil
}

func buildCredentialDescriptors(credentials []orm.WebAuthnCredential) ([]protocol.CredentialDescriptor, error) {
	descriptors := make([]protocol.CredentialDescriptor, 0, len(credentials))
	for _, credential := range credentials {
		id, err := base64.RawURLEncoding.DecodeString(credential.CredentialID)
		if err != nil {
			return nil, err
		}
		transports := make([]protocol.AuthenticatorTransport, 0, len(credential.Transports))
		for _, transport := range credential.Transports {
			transports = append(transports, protocol.AuthenticatorTransport(transport))
		}
		descriptors = append(descriptors, protocol.CredentialDescriptor{
			Type:         protocol.PublicKeyCredentialType,
			CredentialID: id,
			Transport:    transports,
		})
	}
	return descriptors, nil
}

func existsCredential(credentialID string) bool {
	var count int64
	_ = orm.GormDB.Model(&orm.WebAuthnCredential{}).Where("credential_id = ?", credentialID).Count(&count).Error
	return count > 0
}

func credentialExistsForUser(userID string, credentialID string) bool {
	var count int64
	_ = orm.GormDB.Model(&orm.WebAuthnCredential{}).Where("user_id = ? AND credential_id = ?", userID, credentialID).Count(&count).Error
	return count > 0
}

func updateCredentialUsage(credential *webauthn.Credential) error {
	if credential == nil {
		return nil
	}
	credentialID := base64.RawURLEncoding.EncodeToString(credential.ID)
	backupEligible := credential.Flags.BackupEligible
	backupState := credential.Flags.BackupState
	updates := map[string]interface{}{
		"sign_count":      credential.Authenticator.SignCount,
		"last_used_at":    time.Now().UTC(),
		"backup_eligible": &backupEligible,
		"backup_state":    &backupState,
	}
	return orm.GormDB.Model(&orm.WebAuthnCredential{}).Where("credential_id = ?", credentialID).Updates(updates).Error
}
