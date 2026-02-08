package routes

import (
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/handlers"
	"github.com/ggmolly/belfast/internal/api/middleware"
	"github.com/ggmolly/belfast/internal/authz"
)

const apiBasePath = "/api/v1"

func RegisterGameData(app *iris.Application) {
	gameDataHandler := handlers.NewGameDataHandler()
	party := app.Party(apiBasePath)
	party.Use(middleware.RequirePermissionAny(authz.PermGameData))
	handlers.RegisterGameDataRoutes(party, gameDataHandler)
}
