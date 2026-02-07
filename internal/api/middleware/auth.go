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
	authUserKey    = "auth.admin_user"
	authSessionKey = "auth.admin_session"
)

func Auth(cfg *config.Config) iris.Handler {
	authCfg := auth.NormalizeConfig(config.AuthConfig{})
	if cfg != nil {
		authCfg = auth.NormalizeConfig(cfg.Auth)
	}
	cookieName := authCfg.CookieName

	return func(ctx iris.Context) {
		if authCfg.DisableAuth {
			sessionID := ctx.GetCookie(cookieName)
			if sessionID != "" {
				if session, user, err := auth.LoadSession(sessionID); err == nil && user.DisabledAt == nil {
					ctx.Values().Set(authUserKey, user)
					ctx.Values().Set(authSessionKey, session)
				}
			}
			ctx.Next()
			return
		}
		if ctx.Method() == http.MethodOptions {
			ctx.Next()
			return
		}
		if isPublicRoute(ctx.Method(), ctx.Path()) {
			ctx.Next()
			return
		}
		sessionID := ctx.GetCookie(cookieName)
		if sessionID == "" {
			ctx.StatusCode(iris.StatusUnauthorized)
			_ = ctx.JSON(response.Error("auth.session_missing", "session required", nil))
			return
		}
		session, user, err := auth.LoadSession(sessionID)
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
				if err := auth.TouchSession(session.ID, now, newExpires); err == nil {
					session.ExpiresAt = newExpires
					ctx.SetCookie(auth.BuildSessionCookie(authCfg, session))
				}
			} else {
				_ = auth.TouchSession(session.ID, now, time.Time{})
			}
		} else {
			_ = auth.TouchSession(session.ID, now, time.Time{})
		}
		ctx.Values().Set(authUserKey, user)
		ctx.Values().Set(authSessionKey, session)
		ctx.Next()
	}
}

func GetAdminUser(ctx iris.Context) (*orm.AdminUser, bool) {
	user, ok := ctx.Values().Get(authUserKey).(*orm.AdminUser)
	return user, ok
}

func GetAdminSession(ctx iris.Context) (*orm.AdminSession, bool) {
	session, ok := ctx.Values().Get(authSessionKey).(*orm.AdminSession)
	return session, ok
}

var publicRoutePrefixes = []string{
	"/swagger",
	"/api/v1/user/",
	"/api/v1/registration/",
	"/api/v1/me",
}

var publicRouteMethods = map[string]map[string]struct{}{
	"/health": nil,
	"/api/v1/auth/bootstrap": {
		http.MethodPost: {},
	},
	"/api/v1/auth/bootstrap/status": {
		http.MethodGet: {},
	},
	"/api/v1/auth/login": {
		http.MethodPost: {},
	},
	"/api/v1/server/status": {
		http.MethodGet: {},
	},
	"/api/v1/auth/passkeys/authenticate/options": {
		http.MethodPost: {},
	},
	"/api/v1/auth/passkeys/authenticate/verify": {
		http.MethodPost: {},
	},
}

func isPublicRoute(method string, path string) bool {
	for _, prefix := range publicRoutePrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	methods, ok := publicRouteMethods[path]
	if !ok {
		return false
	}
	if methods == nil {
		return true
	}
	_, ok = methods[method]
	return ok
}

func requiresCSRF(method string) bool {
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete
}
