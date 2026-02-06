package routes

import (
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/handlers"
	"github.com/ggmolly/belfast/internal/api/middleware"
	"github.com/ggmolly/belfast/internal/authz"
)

func RegisterAdminPermissionPolicy(app *iris.Application) {
	handler := handlers.NewAdminPermissionPolicyHandler()
	party := app.Party("/api/v1/admin/permission-policy")
	party.Use(middleware.RequirePermissionAny(authz.PermAdminPermission))
	handlers.RegisterAdminPermissionPolicyRoutes(party, handler)
}
