package routes

import (
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/handlers"
)

func RegisterExchangeCodes(app *iris.Application) {
	party := app.Party("/api/v1/exchange-codes")
	handler := handlers.NewExchangeCodeHandler()
	handlers.RegisterExchangeCodeRoutes(party, handler)
}
