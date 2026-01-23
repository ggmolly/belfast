package routes

import (
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/handlers"
)

func RegisterJuustagram(app *iris.Application) {
	handler := handlers.NewJuustagramHandler()
	party := app.Party("/api/v1/juustagram")
	handlers.RegisterJuustagramRoutes(party, handler)
	playerParty := app.Party("/api/v1/players/{id:uint}/juustagram")
	handlers.RegisterJuustagramPlayerRoutes(playerParty, handler)
}
