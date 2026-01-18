package routes

import (
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/handlers"
)

func RegisterPlayers(app *iris.Application) {
	party := app.Party("/api/v1/players")
	handler := handlers.NewPlayerHandler()
	handlers.RegisterPlayerRoutes(party, handler)
}
