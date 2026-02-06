package routes

import (
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/handlers"
	"github.com/ggmolly/belfast/internal/api/middleware"
	"github.com/ggmolly/belfast/internal/authz"
)

func RegisterAdminAuthz(app *iris.Application) {
	handler := handlers.NewAdminAuthzHandler()
	party := app.Party("/api/v1/admin/authz")
	party.Use(middleware.RequirePermissionAny(authz.PermAdminAuthz))
	handlers.RegisterAdminAuthzRoutes(party, handler)
}
