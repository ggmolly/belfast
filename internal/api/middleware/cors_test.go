package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kataras/iris/v12"
)

func TestCORSEmptyOrigins(t *testing.T) {
	app := iris.New()
	app.UseRouter(CORS(nil))
	app.Get("/health", func(ctx iris.Context) {
		ctx.StatusCode(http.StatusOK)
	})

	if err := app.Build(); err != nil {
		t.Fatalf("build app: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
}

func TestCORSWithOrigins(t *testing.T) {
	app := iris.New()
	app.UseRouter(CORS([]string{"http://example.com"}))
	app.Get("/health", func(ctx iris.Context) {
		ctx.StatusCode(http.StatusOK)
	})

	if err := app.Build(); err != nil {
		t.Fatalf("build app: %v", err)
	}

	request := httptest.NewRequest(http.MethodOptions, "/health", nil)
	request.Header.Set("Origin", "http://example.com")
	request.Header.Set("Access-Control-Request-Method", "GET")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Fatalf("expected CORS headers to be set")
	}
}
