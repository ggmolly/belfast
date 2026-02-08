package routes

import (
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/handlers"
	"github.com/ggmolly/belfast/internal/api/middleware"
	"github.com/ggmolly/belfast/internal/authz"
)

func RegisterJuustagram(app *iris.Application) {
	handler := handlers.NewJuustagramHandler()
	adminParty := app.Party("/api/v1/juustagram")
	adminParty.Use(middleware.RequirePermissionAny(authz.PermJuustagram))
	handlers.RegisterJuustagramRoutes(adminParty, handler)
	playerParty := app.Party("/api/v1/players/{id:uint}/juustagram")
	playerParty.Use(middleware.RequirePermissionAnyOrSelf(authz.PermPlayers))
	handlers.RegisterJuustagramPlayerRoutes(playerParty, handler)
}
