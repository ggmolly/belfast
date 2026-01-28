package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kataras/iris/v12"
)

func TestRecoverMiddleware(t *testing.T) {
	app := iris.New()
	app.UseRouter(Recover())
	app.Get("/panic", func(ctx iris.Context) {
		panic("boom")
	})

	if err := app.Build(); err != nil {
		t.Fatalf("build app: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/panic", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", response.Code)
	}
}
