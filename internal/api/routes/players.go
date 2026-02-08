package routes

import (
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/handlers"
	"github.com/ggmolly/belfast/internal/api/middleware"
	"github.com/ggmolly/belfast/internal/authz"
)

func RegisterPlayers(app *iris.Application) {
	party := app.Party("/api/v1/players")
	party.Use(middleware.RequirePermissionAnyOrSelf(authz.PermPlayers))
	handler := handlers.NewPlayerHandler()
	handlers.RegisterPlayerRoutes(party, handler)
}
