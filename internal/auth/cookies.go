package auth

import (
	"net/http"
	"strings"

	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/orm"
)

func BuildSessionCookie(cfg config.AuthConfig, session *orm.Session) *http.Cookie {
	if session == nil {
		return nil
	}
	return &http.Cookie{
		Name:     cfg.CookieName,
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: cookieSameSite(cfg.CookieSameSite),
		Expires:  session.ExpiresAt,
	}
}

func ClearSessionCookie(cfg config.AuthConfig) *http.Cookie {
	return &http.Cookie{
		Name:     cfg.CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: cookieSameSite(cfg.CookieSameSite),
		MaxAge:   -1,
	}
}

func cookieSameSite(value string) http.SameSite {
	switch strings.ToLower(value) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
