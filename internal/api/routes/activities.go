package routes

import (
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/handlers"
	"github.com/ggmolly/belfast/internal/api/middleware"
	"github.com/ggmolly/belfast/internal/authz"
)

func RegisterActivities(app *iris.Application) {
	party := app.Party("/api/v1/activities")
	party.Use(middleware.RequirePermissionAny(authz.PermActivities))
	handler := handlers.NewActivityHandler()
	handlers.RegisterActivityRoutes(party, handler)
}
