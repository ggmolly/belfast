package api

import (
	"fmt"
	"net/http"

	"github.com/kataras/iris/v12"

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
