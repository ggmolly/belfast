package api

import (
	"fmt"
	"net/http"

	"github.com/iris-contrib/swagger"
	"github.com/iris-contrib/swagger/swaggerFiles"
	"github.com/kataras/iris/v12"
	"github.com/swaggo/swag"

	"github.com/ggmolly/belfast/docs"
	"github.com/ggmolly/belfast/internal/api/middleware"
	"github.com/ggmolly/belfast/internal/api/routes"
	"github.com/ggmolly/belfast/internal/logger"
)

func NewApp(cfg Config) *iris.Application {
	app := iris.New()

	app.UseRouter(middleware.Recover())
	app.UseRouter(middleware.RequestLogger())
	app.UseRouter(middleware.CORS(cfg.CORSOrigins))

	middleware.RegisterErrorHandlers(app)
	routes.Register(app)
	routes.RegisterServer(app, cfg.RuntimeConfig)
	routes.RegisterPlayers(app)
	routes.RegisterGameData(app)
	routes.RegisterShop(app)
	routes.RegisterNotices(app)

	swag.Register("doc", docs.SwaggerInfo)
	swaggerUI := swagger.Handler(
		swaggerFiles.Handler,
		swagger.URL("/swagger/doc.json"),
		swagger.Prefix("/swagger"),
	)
	app.Get("/swagger", swaggerUI)
	app.Get("/swagger/{any:path}", swaggerUI)

	return app
}

func Start(cfg Config) error {
	if !cfg.Enabled {
		logger.LogEvent("API", "Start", "API server disabled", logger.LOG_LEVEL_INFO)
		return nil
	}

	app := NewApp(cfg)
	addr := fmt.Sprintf(":%d", cfg.Port)
	server := &http.Server{Addr: addr}
	host := app.NewHost(server)
	logger.LogEvent("API", "Start", fmt.Sprintf("listening on %s", addr), logger.LOG_LEVEL_INFO)
	return app.Run(iris.Raw(host.ListenAndServe))
}
