package routes

import (
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/handlers"
	"github.com/ggmolly/belfast/internal/api/middleware"
	"github.com/ggmolly/belfast/internal/authz"
)

func RegisterShop(app *iris.Application) {
	party := app.Party("/api/v1/shop")
	party.Use(middleware.RequirePermissionAny(authz.PermShop))
	handler := handlers.NewShopHandler()
	handlers.RegisterShopRoutes(party, handler)
}

func RegisterNotices(app *iris.Application) {
	party := app.Party("/api/v1/notices")
	party.Use(middleware.RequirePermissionAny(authz.PermNotices))
	handler := handlers.NewNoticeHandler()
	handlers.RegisterNoticeRoutes(party, handler)
}
