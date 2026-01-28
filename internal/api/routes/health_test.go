package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kataras/iris/v12"
)

func TestHealthRoute(t *testing.T) {
	app := iris.New()
	Register(app)

	if err := app.Build(); err != nil {
		t.Fatalf("build app: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var payload struct {
		OK   bool `json:"ok"`
		Data struct {
			Status string `json:"status"`
			Time   string `json:"time"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !payload.OK || payload.Data.Status != "ok" || payload.Data.Time == "" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}
