package auth

import (
	"strings"
	"time"

	"github.com/ggmolly/belfast/internal/config"
)

const (
	defaultSessionTTLSeconds           = 86400
	defaultCSRFTokenTTLSeconds         = 7200
	defaultWebAuthnChallengeTTLSeconds = 300
	defaultPasswordMinLength           = 12
	defaultPasswordMaxLength           = 128
	defaultRateLimitWindowSeconds      = 60
	defaultRateLimitLoginMax           = 5
	defaultRateLimitPasskeyMax         = 5
	defaultCookieName                  = "belfast_session"
	defaultCookieSameSite              = "lax"
)

var defaultArgon2Params = config.Argon2Config{
	Memory:      65536,
	Iterations:  3,
	Parallelism: 1,
	SaltLength:  16,
	KeyLength:   32,
}

func NormalizeConfig(cfg config.AuthConfig) config.AuthConfig {
	return normalizeConfig(cfg, defaultCookieName)
}

func NormalizeUserConfig(cfg config.AuthConfig) config.AuthConfig {
	return normalizeConfig(cfg, defaultCookieName)
}

func normalizeConfig(cfg config.AuthConfig, defaultCookie string) config.AuthConfig {
	if cfg.SessionTTLSeconds <= 0 {
		cfg.SessionTTLSeconds = defaultSessionTTLSeconds
	}
	if cfg.CookieName == "" {
		cfg.CookieName = defaultCookie
	}
	if cfg.CookieSameSite == "" {
		cfg.CookieSameSite = defaultCookieSameSite
	}
	if cfg.CSRFTokenTTLSeconds <= 0 {
		cfg.CSRFTokenTTLSeconds = defaultCSRFTokenTTLSeconds
	}
	if cfg.PasswordMinLength <= 0 {
		cfg.PasswordMinLength = defaultPasswordMinLength
	}
	if cfg.PasswordMaxLength <= 0 {
		cfg.PasswordMaxLength = defaultPasswordMaxLength
	}
	if cfg.PasswordHashParams.Memory == 0 {
		cfg.PasswordHashParams.Memory = defaultArgon2Params.Memory
	}
	if cfg.PasswordHashParams.Iterations == 0 {
		cfg.PasswordHashParams.Iterations = defaultArgon2Params.Iterations
	}
	if cfg.PasswordHashParams.Parallelism == 0 {
		cfg.PasswordHashParams.Parallelism = defaultArgon2Params.Parallelism
	}
	if cfg.PasswordHashParams.SaltLength == 0 {
		cfg.PasswordHashParams.SaltLength = defaultArgon2Params.SaltLength
	}
	if cfg.PasswordHashParams.KeyLength == 0 {
		cfg.PasswordHashParams.KeyLength = defaultArgon2Params.KeyLength
	}
	if cfg.WebAuthnChallengeTTLSeconds <= 0 {
		cfg.WebAuthnChallengeTTLSeconds = defaultWebAuthnChallengeTTLSeconds
	}
	if len(cfg.WebAuthnExpectedOrigins) > 0 {
		trimmed := make([]string, 0, len(cfg.WebAuthnExpectedOrigins))
		for _, origin := range cfg.WebAuthnExpectedOrigins {
			value := strings.TrimSpace(origin)
			if value == "" {
				continue
			}
			trimmed = append(trimmed, value)
		}
		cfg.WebAuthnExpectedOrigins = trimmed
	}
	if cfg.RateLimitWindowSeconds <= 0 {
		cfg.RateLimitWindowSeconds = defaultRateLimitWindowSeconds
	}
	if cfg.RateLimitLoginMax <= 0 {
		cfg.RateLimitLoginMax = defaultRateLimitLoginMax
	}
	if cfg.RateLimitPasskeyMax <= 0 {
		cfg.RateLimitPasskeyMax = defaultRateLimitPasskeyMax
	}
	return cfg
}

func SessionTTL(cfg config.AuthConfig) time.Duration {
	return time.Duration(cfg.SessionTTLSeconds) * time.Second
}

func CSRFTTL(cfg config.AuthConfig) time.Duration {
	return time.Duration(cfg.CSRFTokenTTLSeconds) * time.Second
}

func WebAuthnChallengeTTL(cfg config.AuthConfig) time.Duration {
	return time.Duration(cfg.WebAuthnChallengeTTLSeconds) * time.Second
}

func RateLimitWindow(cfg config.AuthConfig) time.Duration {
	return time.Duration(cfg.RateLimitWindowSeconds) * time.Second
}
