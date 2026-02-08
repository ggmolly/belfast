package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/middleware"
	"github.com/ggmolly/belfast/internal/api/routes"
	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/connection"
)

func TestServerStatusPublic(t *testing.T) {
	app := iris.New()
	cfg := &config.Config{
		Auth: config.AuthConfig{
			CookieName: "belfast_admin_session",
		},
		Belfast: config.BelfastConfig{
			BindAddress: "127.0.0.1",
			Port:        8080,
			Name:        "Belfast",
		},
	}
	app.UseRouter(middleware.Auth(cfg))
	routes.RegisterServer(app, cfg)
	connection.NewServer("127.0.0.1", 8080, func(pkt *[]byte, c *connection.Client, size int) {})
	if err := app.Build(); err != nil {
		t.Fatalf("build app: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/server/status", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
}
