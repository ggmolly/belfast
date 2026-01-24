package routes

import (
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/handlers"
)

func RegisterActivities(app *iris.Application) {
	party := app.Party("/api/v1/activities")
	handler := handlers.NewActivityHandler()
	handlers.RegisterActivityRoutes(party, handler)
}
