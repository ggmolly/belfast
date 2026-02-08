package api

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/iris-contrib/swagger"
	"github.com/iris-contrib/swagger/swaggerFiles"
	"github.com/kataras/iris/v12"
	"github.com/swaggo/swag"

	"github.com/ggmolly/belfast/docs"
	"github.com/ggmolly/belfast/internal/api/middleware"
	"github.com/ggmolly/belfast/internal/api/routes"
	"github.com/ggmolly/belfast/internal/logger"
)

var swaggerOnce sync.Once

var runServer = func(app *iris.Application, addr string) error {
	server := &http.Server{Addr: addr}
	host := app.NewHost(server)
	return app.Run(iris.Raw(host.ListenAndServe))
}

func NewApp(cfg Config) *iris.Application {
	app := iris.New()

	app.UseRouter(middleware.Recover())
	app.UseRouter(middleware.RequestLogger())
	app.UseRouter(middleware.CORS(cfg.CORSOrigins))
	app.UseRouter(middleware.Auth(cfg.RuntimeConfig))
	app.UseRouter(middleware.Audit())

	middleware.RegisterErrorHandlers(app)
	routes.Register(app)
	authManager := routes.RegisterAuth(app, cfg.RuntimeConfig)
	routes.RegisterAdminAuthz(app)
	routes.RegisterAdminUsers(app, authManager)
	routes.RegisterAdminPermissionPolicy(app)
	routes.RegisterRegistration(app, cfg.RuntimeConfig)
	routes.RegisterUserAuth(app, cfg.RuntimeConfig)
	routes.RegisterMe(app, cfg.RuntimeConfig)
	routes.RegisterServer(app, cfg.RuntimeConfig)
	routes.RegisterPlayers(app)
	routes.RegisterGameData(app)
	routes.RegisterShop(app)
	routes.RegisterNotices(app)
	routes.RegisterExchangeCodes(app)
	routes.RegisterDorm3d(app)
	routes.RegisterJuustagram(app)
	routes.RegisterActivities(app)

	swaggerOnce.Do(func() {
		swag.Register("doc", docs.SwaggerInfo)
	})
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
	logger.LogEvent("API", "Start", fmt.Sprintf("listening on %s", addr), logger.LOG_LEVEL_INFO)
	return runServer(app, addr)
}
