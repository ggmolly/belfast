package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ggmolly/belfast/internal/config"
	"github.com/kataras/iris/v12"
)

func TestStartDisabled(t *testing.T) {
	if err := Start(Config{Enabled: false}); err != nil {
		t.Fatalf("expected nil when disabled, got %v", err)
	}
}

func TestStartEnabledUsesRunner(t *testing.T) {
	original := runServer
	defer func() {
		runServer = original
	}()

	called := false
	var gotAddr string
	runServer = func(app *iris.Application, addr string) error {
		called = true
		gotAddr = addr
		if app == nil {
			t.Fatalf("expected app")
		}
		return nil
	}

	cfg := Config{Enabled: true, Port: 4242, RuntimeConfig: &config.Config{}}
	if err := Start(cfg); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if !called {
		t.Fatalf("expected runServer to be called")
	}
	if gotAddr != ":4242" {
		t.Fatalf("expected addr :4242, got %s", gotAddr)
	}
}

func TestNewAppRegistersRoutesAndMiddleware(t *testing.T) {
	cfg := Config{CORSOrigins: []string{"http://example.com"}, RuntimeConfig: &config.Config{}}
	app := NewApp(cfg)
	app.Build()

	preflight := httptest.NewRequest(http.MethodOptions, "/health", nil)
	preflight.Header.Set("Origin", "http://example.com")
	preflight.Header.Set("Access-Control-Request-Method", http.MethodGet)
	preflightResponse := httptest.NewRecorder()
	app.ServeHTTP(preflightResponse, preflight)
	if preflightResponse.Code != http.StatusOK && preflightResponse.Code != http.StatusNoContent {
		t.Fatalf("expected preflight 200/204, got %d", preflightResponse.Code)
	}
	allowOrigin := preflightResponse.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "http://example.com" && allowOrigin != "*" {
		t.Fatalf("expected allow origin header, got %q", allowOrigin)
	}

	healthRequest := httptest.NewRequest(http.MethodGet, "/health", nil)
	healthResponse := httptest.NewRecorder()
	app.ServeHTTP(healthResponse, healthRequest)
	if healthResponse.Code != http.StatusOK {
		t.Fatalf("expected 200 for health, got %d", healthResponse.Code)
	}

	swaggerRequest := httptest.NewRequest(http.MethodGet, "/swagger", nil)
	swaggerResponse := httptest.NewRecorder()
	app.ServeHTTP(swaggerResponse, swaggerRequest)
	swaggerStatus := swaggerResponse.Code
	if swaggerStatus != http.StatusOK && swaggerStatus != http.StatusMovedPermanently && swaggerStatus != http.StatusFound {
		t.Fatalf("expected swagger route, got %d", swaggerStatus)
	}
}

func TestRunServerInvalidAddrReturnsError(t *testing.T) {
	app := iris.New()
	if err := runServer(app, "127.0.0.1"); err == nil {
		t.Fatalf("expected error")
	}
}
