package routes

import (
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/handlers"
	"github.com/ggmolly/belfast/internal/api/middleware"
	"github.com/ggmolly/belfast/internal/auth"
	"github.com/ggmolly/belfast/internal/authz"
)

func RegisterAdminUsers(app *iris.Application, manager *auth.Manager) {
	handler := handlers.NewAdminUserHandler(manager)
	party := app.Party("/api/v1/admin/users")
	party.Use(middleware.RequirePermissionAny(authz.PermAdminUsers))
	handlers.RegisterAdminUserRoutes(party, handler)
}
