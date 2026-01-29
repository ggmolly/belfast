package routes

import (
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/handlers"
)

const apiBasePath = "/api/v1"

func RegisterGameData(app *iris.Application) {
	gameDataHandler := handlers.NewGameDataHandler()
	handlers.RegisterGameDataRoutes(app.Party(apiBasePath), gameDataHandler)
}
