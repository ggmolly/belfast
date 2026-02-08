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
	authAccountKey  = "auth.account"
	authSessionKey  = "auth.session"
	authDisabledKey = "auth.disabled"
)

func Auth(cfg *config.Config) iris.Handler {
	authCfg := auth.NormalizeConfig(config.AuthConfig{})
	if cfg != nil {
		authCfg = auth.NormalizeConfig(cfg.Auth)
	}
	cookieName := authCfg.CookieName

	return func(ctx iris.Context) {
		ctx.Values().Set(authDisabledKey, authCfg.DisableAuth)
		if authCfg.DisableAuth {
			sessionID := ctx.GetCookie(cookieName)
			if sessionID != "" {
				if session, account, err := auth.LoadSession(sessionID); err == nil && account.DisabledAt == nil {
					ctx.Values().Set(authAccountKey, account)
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
		session, account, err := auth.LoadSession(sessionID)
		if err != nil {
			ctx.StatusCode(iris.StatusUnauthorized)
			_ = ctx.JSON(response.Error("auth.session_missing", "session required", nil))
			return
		}
		if account.DisabledAt != nil {
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
		ctx.Values().Set(authAccountKey, account)
		ctx.Values().Set(authSessionKey, session)
		ctx.Next()
	}
}

func IsAuthDisabled(ctx iris.Context) bool {
	disabled, ok := ctx.Values().Get(authDisabledKey).(bool)
	return ok && disabled
}

func GetAccount(ctx iris.Context) (*orm.Account, bool) {
	account, ok := ctx.Values().Get(authAccountKey).(*orm.Account)
	return account, ok
}

func GetSession(ctx iris.Context) (*orm.Session, bool) {
	session, ok := ctx.Values().Get(authSessionKey).(*orm.Session)
	return session, ok
}

var publicRoutePrefixes = []string{
	"/swagger",
	"/api/v1/registration/",
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
	"/api/v1/user/auth/login": {
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
