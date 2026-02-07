package routes

import (
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/handlers"
	"github.com/ggmolly/belfast/internal/config"
)

func RegisterRegistration(app *iris.Application, cfg *config.Config) {
	handler := handlers.NewRegistrationHandler(cfg)
	party := app.Party("/api/v1/registration")
	handlers.RegisterRegistrationRoutes(party, handler)
}
