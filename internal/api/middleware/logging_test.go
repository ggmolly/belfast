package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kataras/iris/v12"
)

func TestRequestLogger(t *testing.T) {
	app := iris.New()
	app.UseRouter(RequestLogger())
	app.Get("/ping", func(ctx iris.Context) {
		ctx.StatusCode(http.StatusOK)
	})

	if err := app.Build(); err != nil {
		t.Fatalf("build app: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
}
