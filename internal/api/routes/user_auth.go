package routes

import (
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/handlers"
	"github.com/ggmolly/belfast/internal/config"
)

func RegisterUserAuth(app *iris.Application, cfg *config.Config) {
	handler := handlers.NewUserAuthHandler(cfg)
	party := app.Party("/api/v1/user/auth")
	handlers.RegisterUserAuthRoutes(party, handler)
}
