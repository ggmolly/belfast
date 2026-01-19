package routes

import (
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/handlers"
)

func RegisterGameData(app *iris.Application) {
	party := app.Party("/api/v1")
	handler := handlers.NewGameDataHandler()
	handlers.RegisterGameDataRoutes(party, handler)
}
