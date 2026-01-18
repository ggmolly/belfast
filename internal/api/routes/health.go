package routes

import (
	"time"

	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/response"
)

type healthResponse struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

func Register(app *iris.Application) {
	app.Get("/health", func(ctx iris.Context) {
		payload := healthResponse{
			Status: "ok",
			Time:   time.Now().UTC().Format(time.RFC3339),
		}
		_ = ctx.JSON(response.Success(payload))
	})

	api := app.Party("/api/v1")
	{
		_ = api
	}
}
