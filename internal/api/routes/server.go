package routes

import (
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/handlers"
	"github.com/ggmolly/belfast/internal/config"
)

func RegisterServer(app *iris.Application, cfg *config.Config) {
	party := app.Party("/api/v1/server")
	handler := &handlers.ServerHandler{Config: cfg}
	handlers.RegisterServerRoutes(party, handler)
}
