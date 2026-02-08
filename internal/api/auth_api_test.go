package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/middleware"
	"github.com/ggmolly/belfast/internal/api/response"
	"github.com/ggmolly/belfast/internal/api/routes"
	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/orm"
)

var authTestOnce sync.Once

func initAuthTestDB(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	authTestOnce.Do(func() {
		orm.InitDatabase()
	})
}

func clearAuthTables(t *testing.T) {
	t.Helper()
	tables := []string{"audit_logs", "auth_challenges", "web_authn_credentials", "account_permission_overrides", "account_roles", "sessions", "accounts"}
	for _, table := range tables {
		if err := orm.GormDB.Exec("DELETE FROM " + table).Error; err != nil {
			t.Fatalf("clear %s: %v", table, err)
		}
	}
}

func newAuthTestApp(t *testing.T) *iris.Application {
	initAuthTestDB(t)
	app := iris.New()
	cfg := &config.Config{
		Auth: config.AuthConfig{
			CookieName:              "belfast_session",
			WebAuthnRPID:            "localhost",
			WebAuthnRPName:          "Belfast Admin",
			WebAuthnExpectedOrigins: []string{"http://localhost"},
		},
	}
	app.UseRouter(middleware.Auth(cfg))
	manager := routes.RegisterAuth(app, cfg)
	routes.RegisterAdminUsers(app, manager)
	app.Get("/api/v1/protected", func(ctx iris.Context) {
		_ = ctx.JSON(response.Success(map[string]string{"status": "ok"}))
	})
	if err := app.Build(); err != nil {
		t.Fatalf("build app: %v", err)
	}
	return app
}

func TestAuthBootstrapStatus(t *testing.T) {
	app := newAuthTestApp(t)
	clearAuthTables(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/bootstrap/status", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	var statusPayload struct {
		OK   bool `json:"ok"`
		Data struct {
			CanBootstrap bool  `json:"can_bootstrap"`
			AdminCount   int64 `json:"admin_count"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&statusPayload); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	if !statusPayload.Data.CanBootstrap {
		t.Fatalf("expected can_bootstrap true")
	}
	if statusPayload.Data.AdminCount != 0 {
		t.Fatalf("expected admin_count 0, got %d", statusPayload.Data.AdminCount)
	}

	bootstrap := `{"username":"admin","password":"this-is-a-strong-pass"}`
	request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/bootstrap", strings.NewReader(bootstrap))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected bootstrap 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/bootstrap/status", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if err := json.NewDecoder(response.Body).Decode(&statusPayload); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	if statusPayload.Data.CanBootstrap {
		t.Fatalf("expected can_bootstrap false")
	}
	if statusPayload.Data.AdminCount != 1 {
		t.Fatalf("expected admin_count 1, got %d", statusPayload.Data.AdminCount)
	}
}

func TestAuthBootstrapAndLogin(t *testing.T) {
	app := newAuthTestApp(t)
	clearAuthTables(t)

	bootstrap := `{"username":"admin","password":"this-is-a-strong-pass"}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/bootstrap", strings.NewReader(bootstrap))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected bootstrap 200, got %d", response.Code)
	}
	if len(response.Result().Cookies()) == 0 {
		t.Fatalf("expected session cookie")
	}

	request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/bootstrap", strings.NewReader(bootstrap))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusConflict {
		t.Fatalf("expected bootstrap conflict, got %d", response.Code)
	}

	badLogin := `{"username":"admin","password":"wrong-password"}`
	request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(badLogin))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected login 401, got %d", response.Code)
	}

	goodLogin := `{"username":"admin","password":"this-is-a-strong-pass"}`
	request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(goodLogin))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d", response.Code)
	}
}

func TestAuthSessionAndProtectedRoute(t *testing.T) {
	app := newAuthTestApp(t)
	clearAuthTables(t)

	bootstrap := `{"username":"admin","password":"this-is-a-strong-pass"}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/bootstrap", strings.NewReader(bootstrap))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected bootstrap 200, got %d", response.Code)
	}
	var bootstrapResponse struct {
		OK   bool `json:"ok"`
		Data struct {
			User struct {
				ID string `json:"id"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&bootstrapResponse); err != nil {
		t.Fatalf("decode bootstrap: %v", err)
	}
	if bootstrapResponse.Data.User.ID == "" {
		t.Fatalf("expected bootstrap user id")
	}
	cookies := response.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected session cookie")
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/session", nil)
	request.AddCookie(cookies[0])
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected session 200, got %d", response.Code)
	}
	var sessionPayload struct {
		OK   bool `json:"ok"`
		Data struct {
			CSRFToken string `json:"csrf_token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&sessionPayload); err != nil {
		t.Fatalf("decode session: %v", err)
	}
	if sessionPayload.Data.CSRFToken == "" {
		t.Fatalf("expected csrf token")
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/protected", nil)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected protected 401, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/protected", nil)
	request.AddCookie(cookies[0])
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected protected 200, got %d", response.Code)
	}
}

func TestAdminUserLifecycle(t *testing.T) {
	app := newAuthTestApp(t)
	clearAuthTables(t)

	bootstrap := `{"username":"admin","password":"this-is-a-strong-pass"}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/bootstrap", strings.NewReader(bootstrap))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected bootstrap 200, got %d", response.Code)
	}
	var bootstrapResponse struct {
		OK   bool `json:"ok"`
		Data struct {
			User struct {
				ID string `json:"id"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&bootstrapResponse); err != nil {
		t.Fatalf("decode bootstrap: %v", err)
	}
	if bootstrapResponse.Data.User.ID == "" {
		t.Fatalf("expected bootstrap user id")
	}
	cookies := response.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected session cookie")
	}

	csrfToken := fetchCSRFToken(t, app, cookies[0])
	createPayload := `{"username":"second","password":"this-is-a-strong-pass"}`
	request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", strings.NewReader(createPayload))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", csrfToken)
	request.AddCookie(cookies[0])
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected create 200, got %d", response.Code)
	}
	var createResponse struct {
		OK   bool `json:"ok"`
		Data struct {
			User struct {
				ID string `json:"id"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&createResponse); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	if createResponse.Data.User.ID == "" {
		t.Fatalf("expected user id")
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	request.AddCookie(cookies[0])
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected list 200, got %d", response.Code)
	}

	disablePayload := `{"disabled":true}`
	request = httptest.NewRequest(http.MethodPatch, "/api/v1/admin/users/"+createResponse.Data.User.ID, strings.NewReader(disablePayload))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", csrfToken)
	request.AddCookie(cookies[0])
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected disable 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodPatch, "/api/v1/admin/users/"+bootstrapResponse.Data.User.ID, strings.NewReader(disablePayload))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", csrfToken)
	request.AddCookie(cookies[0])
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusConflict {
		t.Fatalf("expected last admin conflict, got %d", response.Code)
	}
}

func TestPasswordChange(t *testing.T) {
	app := newAuthTestApp(t)
	clearAuthTables(t)

	bootstrap := `{"username":"admin","password":"this-is-a-strong-pass"}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/bootstrap", strings.NewReader(bootstrap))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected bootstrap 200, got %d", response.Code)
	}
	cookies := response.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected session cookie")
	}
	csrfToken := fetchCSRFToken(t, app, cookies[0])

	change := `{"current_password":"this-is-a-strong-pass","new_password":"this-is-a-new-pass"}`
	request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/password/change", strings.NewReader(change))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", csrfToken)
	request.AddCookie(cookies[0])
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected change 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"admin","password":"this-is-a-strong-pass"}`))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected old password 401, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"admin","password":"this-is-a-new-pass"}`))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected new password login 200, got %d", response.Code)
	}
}

func TestPasskeyOptionsAndVerifyFailure(t *testing.T) {
	app := newAuthTestApp(t)
	clearAuthTables(t)

	bootstrap := `{"username":"admin","password":"this-is-a-strong-pass"}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/bootstrap", strings.NewReader(bootstrap))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected bootstrap 200, got %d", response.Code)
	}
	cookies := response.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected session cookie")
	}
	csrfToken := fetchCSRFToken(t, app, cookies[0])

	request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/passkeys/register/options", bytes.NewBufferString(`{}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", csrfToken)
	request.AddCookie(cookies[0])
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected passkey options 200, got %d", response.Code)
	}

	verifyPayload := `{"credential":{"id":"AA","rawId":"AA","type":"public-key","response":{"clientDataJSON":"AA","attestationObject":"AA"}}}`
	request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/passkeys/register/verify", strings.NewReader(verifyPayload))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", csrfToken)
	request.AddCookie(cookies[0])
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code == http.StatusOK {
		t.Fatalf("expected passkey verify failure")
	}
}

func fetchCSRFToken(t *testing.T, app *iris.Application, cookie *http.Cookie) string {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/session", nil)
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected session 200, got %d", response.Code)
	}
	var payload struct {
		OK   bool `json:"ok"`
		Data struct {
			CSRFToken string `json:"csrf_token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode session: %v", err)
	}
	if payload.Data.CSRFToken == "" {
		t.Fatalf("expected csrf token")
	}
	return payload.Data.CSRFToken
}
