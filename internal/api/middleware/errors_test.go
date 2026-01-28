package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kataras/iris/v12"
)

func TestErrorCode(t *testing.T) {
	cases := []struct {
		status   int
		expected string
	}{
		{iris.StatusBadRequest, "bad_request"},
		{iris.StatusUnauthorized, "unauthorized"},
		{iris.StatusForbidden, "forbidden"},
		{iris.StatusNotFound, "not_found"},
		{iris.StatusTooManyRequests, "rate_limited"},
		{iris.StatusInternalServerError, "internal"},
		{iris.StatusNotImplemented, "not_implemented"},
	}
	for _, tt := range cases {
		if code := errorCode(tt.status); code != tt.expected {
			t.Fatalf("expected %s for %d, got %s", tt.expected, tt.status, code)
		}
	}
}

func TestRegisterErrorHandlers(t *testing.T) {
	app := iris.New()
	RegisterErrorHandlers(app)
	app.Get("/boom", func(ctx iris.Context) {
		ctx.StopWithError(iris.StatusBadRequest, iris.NewProblem().Title("bad request"))
	})

	if err := app.Build(); err != nil {
		t.Fatalf("build app: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/boom", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestRegisterErrorHandlersUsesContextError(t *testing.T) {
	app := iris.New()
	RegisterErrorHandlers(app)
	app.Get("/error", func(ctx iris.Context) {
		ctx.SetErr(errors.New("custom error"))
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.StopExecution()
	})

	if err := app.Build(); err != nil {
		t.Fatalf("build app: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/error", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "custom error") {
		t.Fatalf("expected response to contain custom error")
	}
}

func TestRegisterErrorHandlersUsesStatusText(t *testing.T) {
	app := iris.New()
	RegisterErrorHandlers(app)
	app.Get("/missing", func(ctx iris.Context) {
		ctx.StopWithStatus(iris.StatusNotFound)
	})

	if err := app.Build(); err != nil {
		t.Fatalf("build app: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/missing", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "not_found") {
		t.Fatalf("expected response to contain error code")
	}
}
