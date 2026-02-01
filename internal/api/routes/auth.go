package routes

import (
	"github.com/kataras/iris/v12"

	"github.com/go-webauthn/webauthn/protocol"

	"github.com/ggmolly/belfast/internal/api/handlers"
	"github.com/ggmolly/belfast/internal/auth"
	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/logger"
)

func RegisterAuth(app *iris.Application, cfg *config.Config) *auth.Manager {
	if cfg == nil {
		cfg = &config.Config{}
	}
	manager, err := auth.NewManager(cfg.Auth)
	if err != nil {
		logger.LogEvent("API", "Auth", "webauthn disabled: "+err.Error(), logger.LOG_LEVEL_WARN)
		manager = &auth.Manager{Config: auth.NormalizeConfig(cfg.Auth), Limiter: auth.NewRateLimiter(), Selection: protocol.AuthenticatorSelection{UserVerification: protocol.VerificationPreferred}}
	}
	handler := handlers.NewAuthHandler(manager)
	party := app.Party("/api/v1/auth")
	handlers.RegisterAuthRoutes(party, handler)
	return manager
}
