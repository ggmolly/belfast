package routes

import (
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/handlers"
	"github.com/ggmolly/belfast/internal/api/middleware"
	"github.com/ggmolly/belfast/internal/config"
)

func RegisterMe(app *iris.Application, cfg *config.Config) {
	handler := handlers.NewMeHandler()
	party := app.Party("/api/v1/me")
	party.Use(middleware.UserAuth(cfg))
	handlers.RegisterMeRoutes(party, handler)
}
