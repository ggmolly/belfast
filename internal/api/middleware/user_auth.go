package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/response"
	"github.com/ggmolly/belfast/internal/auth"
	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/orm"
)

const (
	userAuthUserKey    = "auth.user_account"
	userAuthSessionKey = "auth.user_session"
)

func UserAuth(cfg *config.Config) iris.Handler {
	authCfg := auth.NormalizeUserConfig(config.AuthConfig{})
	if cfg != nil {
		authCfg = auth.NormalizeUserConfig(cfg.UserAuth)
	}
	cookieName := authCfg.CookieName

	return func(ctx iris.Context) {
		if authCfg.DisableAuth {
			sessionID := ctx.GetCookie(cookieName)
			if sessionID != "" {
				if session, user, err := auth.LoadUserSession(sessionID); err == nil && user.DisabledAt == nil {
					ctx.Values().Set(userAuthUserKey, user)
					ctx.Values().Set(userAuthSessionKey, session)
				}
			}
			ctx.Next()
			return
		}
		if ctx.Method() == http.MethodOptions {
			ctx.Next()
			return
		}
		if isPublicUserRoute(ctx.Method(), ctx.Path()) {
			ctx.Next()
			return
		}
		sessionID := ctx.GetCookie(cookieName)
		if sessionID == "" {
			ctx.StatusCode(iris.StatusUnauthorized)
			_ = ctx.JSON(response.Error("auth.session_missing", "session required", nil))
			return
		}
		session, user, err := auth.LoadUserSession(sessionID)
		if err != nil {
			ctx.StatusCode(iris.StatusUnauthorized)
			_ = ctx.JSON(response.Error("auth.session_missing", "session required", nil))
			return
		}
		if user.DisabledAt != nil {
			ctx.StatusCode(iris.StatusForbidden)
			_ = ctx.JSON(response.Error("auth.user_disabled", "user disabled", nil))
			return
		}
		if requiresCSRF(ctx.Method()) {
			token := ctx.GetHeader("X-CSRF-Token")
			if token == "" || token != session.CSRFToken || session.CSRFExpiresAt.Before(time.Now().UTC()) {
				ctx.StatusCode(iris.StatusForbidden)
				_ = ctx.JSON(response.Error("auth.csrf_invalid", "csrf token required", nil))
				return
			}
		}
		now := time.Now().UTC()
		if authCfg.SessionSliding {
			newExpires := now.Add(auth.SessionTTL(authCfg))
			if newExpires.After(session.ExpiresAt) {
				if err := auth.TouchUserSession(session.ID, now, newExpires); err == nil {
					session.ExpiresAt = newExpires
					ctx.SetCookie(auth.BuildUserSessionCookie(authCfg, session))
				}
			} else {
				_ = auth.TouchUserSession(session.ID, now, time.Time{})
			}
		} else {
			_ = auth.TouchUserSession(session.ID, now, time.Time{})
		}
		ctx.Values().Set(userAuthUserKey, user)
		ctx.Values().Set(userAuthSessionKey, session)
		ctx.Next()
	}
}

func GetUserAccount(ctx iris.Context) (*orm.UserAccount, bool) {
	user, ok := ctx.Values().Get(userAuthUserKey).(*orm.UserAccount)
	return user, ok
}

func GetUserSession(ctx iris.Context) (*orm.UserSession, bool) {
	session, ok := ctx.Values().Get(userAuthSessionKey).(*orm.UserSession)
	return session, ok
}

func isPublicUserRoute(method string, path string) bool {
	if method == http.MethodPost && path == "/api/v1/user/auth/login" {
		return true
	}
	if strings.HasPrefix(path, "/api/v1/registration/") {
		return true
	}
	return false
}
