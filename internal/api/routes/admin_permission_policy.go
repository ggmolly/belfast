package routes

import (
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/handlers"
)

func RegisterAdminPermissionPolicy(app *iris.Application) {
	handler := handlers.NewAdminPermissionPolicyHandler()
	party := app.Party("/api/v1/admin/permission-policy")
	handlers.RegisterAdminPermissionPolicyRoutes(party, handler)
}
