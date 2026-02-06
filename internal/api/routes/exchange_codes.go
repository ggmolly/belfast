package routes

import (
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/handlers"
	"github.com/ggmolly/belfast/internal/api/middleware"
	"github.com/ggmolly/belfast/internal/authz"
)

func RegisterExchangeCodes(app *iris.Application) {
	party := app.Party("/api/v1/exchange-codes")
	party.Use(middleware.RequirePermissionAny(authz.PermExchangeCodes))
	handler := handlers.NewExchangeCodeHandler()
	handlers.RegisterExchangeCodeRoutes(party, handler)
}
