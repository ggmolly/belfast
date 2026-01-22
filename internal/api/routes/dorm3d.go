package routes

import (
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/handlers"
)

func RegisterDorm3d(app *iris.Application) {
	party := app.Party("/api/v1/dorm3d-apartments")
	handler := handlers.NewDorm3dHandler()
	handlers.RegisterDorm3dRoutes(party, handler)
}
