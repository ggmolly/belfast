package routes

import (
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/handlers"
)

func RegisterShop(app *iris.Application) {
	party := app.Party("/api/v1/shop")
	handler := handlers.NewShopHandler()
	handlers.RegisterShopRoutes(party, handler)
}

func RegisterNotices(app *iris.Application) {
	party := app.Party("/api/v1/notices")
	handler := handlers.NewNoticeHandler()
	handlers.RegisterNoticeRoutes(party, handler)
}
