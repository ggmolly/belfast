package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/middleware"
	"github.com/ggmolly/belfast/internal/api/routes"
	"github.com/ggmolly/belfast/internal/auth"
	"github.com/ggmolly/belfast/internal/authz"
	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
)

func newPlayerAuthzTestApp(t *testing.T) *iris.Application {
	t.Helper()
	initAuthTestDB(t)
	app := iris.New()
	cfg := &config.Config{Auth: config.AuthConfig{CookieName: "belfast_session"}}
	app.UseRouter(middleware.Auth(cfg))
	routes.RegisterUserAuth(app, cfg)
	routes.RegisterPlayers(app)
	if err := app.Build(); err != nil {
		t.Fatalf("build app: %v", err)
	}
	return app
}

func clearPlayerAuthzTables(t *testing.T) {
	t.Helper()
	tables := []string{
		"audit_logs",
		"sessions",
		"account_permission_overrides",
		"account_roles",
		"accounts",
		"punishments",
		"owned_resources",
		"commander_items",
		"commander_misc_items",
		"owned_skins",
		"owned_ships",
		"ships",
		"skins",
		"resources",
		"items",
		"commanders",
	}
	for _, table := range tables {
		execAPITestSQLT(t, fmt.Sprintf("DELETE FROM %s", table))
	}
	execAPITestSQLT(t, "DELETE FROM role_permissions WHERE role_id = (SELECT id FROM roles WHERE name = $1)", authz.RolePlayer)
}

func createPlayerAuthzCommander(t *testing.T, commanderID uint32, name string) {
	t.Helper()
	if err := orm.CreateCommanderRoot(commanderID, commanderID, name, 0, 0); err != nil {
		t.Fatalf("create commander: %v", err)
	}
}

func createPlayerAccount(t *testing.T, commanderID uint32, password string) orm.Account {
	t.Helper()
	passwordHash, algo, err := auth.HashPassword(password, auth.NormalizeConfig(config.AuthConfig{}))
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	account := orm.Account{
		ID:                uuid.NewString(),
		CommanderID:       &commanderID,
		PasswordHash:      passwordHash,
		PasswordAlgo:      algo,
		PasswordUpdatedAt: time.Now().UTC(),
	}
	if err := orm.CreateAccount(&account); err != nil {
		t.Fatalf("create account: %v", err)
	}
	if err := orm.AssignRoleByName(account.ID, authz.RolePlayer); err != nil {
		t.Fatalf("assign role: %v", err)
	}
	return account
}

func loginPlayer(t *testing.T, app *iris.Application, commanderID uint32, password string) *http.Cookie {
	t.Helper()
	body := []byte(fmt.Sprintf(`{"commander_id":%d,"password":"%s"}`, commanderID, password))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/user/auth/login", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d", response.Code)
	}
	cookies := response.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected session cookie")
	}
	return cookies[0]
}

func fetchUserCSRFToken(t *testing.T, app *iris.Application, cookie *http.Cookie) string {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/user/auth/session", nil)
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

func TestPlayersAPIReadSelfAllowsOwnCommander(t *testing.T) {
	app := newPlayerAuthzTestApp(t)
	clearPlayerAuthzTables(t)
	if err := orm.EnsureAuthzDefaults(); err != nil {
		t.Fatalf("ensure authz defaults: %v", err)
	}

	password := "this-is-a-strong-pass"

	createPlayerAuthzCommander(t, 111, "Self")
	createPlayerAuthzCommander(t, 222, "Other")

	createPlayerAccount(t, 111, password)
	if err := orm.ReplaceRolePolicyByName(authz.RolePlayer, map[string]authz.Capability{authz.PermPlayers: {ReadSelf: true}}, nil); err != nil {
		t.Fatalf("update policy: %v", err)
	}
	cookie := loginPlayer(t, app, 111, password)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players/111", nil)
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected detail 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/players/111/items", nil)
	request.AddCookie(cookie)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected items 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/players/222/items", nil)
	request.AddCookie(cookie)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("expected other items 403, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/players?offset=0&limit=1", nil)
	request.AddCookie(cookie)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("expected list 403, got %d", response.Code)
	}

	csrfToken := fetchUserCSRFToken(t, app, cookie)
	patchBody := []byte(`{"level":12}`)
	request = httptest.NewRequest(http.MethodPatch, "/api/v1/players/111", bytes.NewReader(patchBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", csrfToken)
	request.AddCookie(cookie)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("expected patch 403 (no write_self), got %d", response.Code)
	}
}

func TestPlayersAPIWriteSelfAllowsOwnCommanderMutations(t *testing.T) {
	app := newPlayerAuthzTestApp(t)
	clearPlayerAuthzTables(t)
	if err := orm.EnsureAuthzDefaults(); err != nil {
		t.Fatalf("ensure authz defaults: %v", err)
	}

	password := "this-is-a-strong-pass"
	createPlayerAuthzCommander(t, 111, "Self")
	createPlayerAuthzCommander(t, 222, "Other")

	createPlayerAccount(t, 111, password)
	if err := orm.ReplaceRolePolicyByName(authz.RolePlayer, map[string]authz.Capability{authz.PermPlayers: {WriteSelf: true}}, nil); err != nil {
		t.Fatalf("update policy: %v", err)
	}
	cookie := loginPlayer(t, app, 111, password)
	csrfToken := fetchUserCSRFToken(t, app, cookie)

	// Seed minimal data for write endpoints.
	execAPITestSQLT(t, "INSERT INTO items (id, name, rarity, shop_id, type, virtual_type) VALUES ($1, $2, $3, $4, $5, $6)", int64(1), "Resource Item", int64(1), int64(-2), int64(0), int64(0))
	execAPITestSQLT(t, "INSERT INTO resources (id, item_id, name) VALUES ($1, $2, $3)", int64(1), int64(1), "Gold")
	execAPITestSQLT(t, "INSERT INTO items (id, name, rarity, shop_id, type, virtual_type) VALUES ($1, $2, $3, $4, $5, $6)", int64(20001), "Test Item", int64(1), int64(-2), int64(0), int64(0))
	execAPITestSQLT(t, "INSERT INTO ships (template_id, name, english_name, rarity_id, star, type, nationality, build_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", int64(1), "Ship", "Ship", int64(2), int64(1), int64(1), int64(1), int64(1))
	now := time.Now().UTC()
	execAPITestSQLT(t, "INSERT INTO owned_ships (id, owner_id, ship_id, create_time, change_name_timestamp) VALUES ($1, $2, $3, $4, $5)", int64(8001), int64(111), int64(1), now, now)
	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	execAPITestSQLT(t, "INSERT INTO skins (id, name, ship_group) VALUES ($1, $2, $3)", int64(1002), "Skin", int64(1))
	execAPITestSQLT(t, "INSERT INTO owned_skins (commander_id, skin_id, expires_at) VALUES ($1, $2, $3)", int64(111), int64(1002), expiresAt)

	patchBody := []byte(`{"level":12}`)
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/players/111", bytes.NewReader(patchBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", csrfToken)
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected patch self 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodPatch, "/api/v1/players/222", bytes.NewReader(patchBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", csrfToken)
	request.AddCookie(cookie)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("expected patch other 403, got %d", response.Code)
	}

	resourceBody := []byte(`{"resources":[{"resource_id":1,"amount":500}]}`)
	request = httptest.NewRequest(http.MethodPut, "/api/v1/players/111/resources", bytes.NewReader(resourceBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", csrfToken)
	request.AddCookie(cookie)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected update resources 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodPut, "/api/v1/players/222/resources", bytes.NewReader(resourceBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", csrfToken)
	request.AddCookie(cookie)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("expected update other resources 403, got %d", response.Code)
	}

	itemBody := []byte(`{"quantity":3}`)
	request = httptest.NewRequest(http.MethodPatch, "/api/v1/players/111/items/20001", bytes.NewReader(itemBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", csrfToken)
	request.AddCookie(cookie)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected update item 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodPatch, "/api/v1/players/222/items/20001", bytes.NewReader(itemBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", csrfToken)
	request.AddCookie(cookie)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("expected update other item 403, got %d", response.Code)
	}

	shipBody := []byte(`{"level":2}`)
	request = httptest.NewRequest(http.MethodPatch, "/api/v1/players/111/ships/8001", bytes.NewReader(shipBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", csrfToken)
	request.AddCookie(cookie)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected update ship 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodPatch, "/api/v1/players/222/ships/8001", bytes.NewReader(shipBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", csrfToken)
	request.AddCookie(cookie)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("expected update other ship 403, got %d", response.Code)
	}

	skinBody := []byte(`{"expires_at":"2030-01-01T00:00:00Z"}`)
	request = httptest.NewRequest(http.MethodPatch, "/api/v1/players/111/skins/1002", bytes.NewReader(skinBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", csrfToken)
	request.AddCookie(cookie)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected update skin 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodPatch, "/api/v1/players/222/skins/1002", bytes.NewReader(skinBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", csrfToken)
	request.AddCookie(cookie)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("expected update other skin 403, got %d", response.Code)
	}

	banBody := []byte(`{"duration_sec":60}`)
	request = httptest.NewRequest(http.MethodPost, "/api/v1/players/111/ban", bytes.NewReader(banBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", csrfToken)
	request.AddCookie(cookie)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected ban self 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodPost, "/api/v1/players/222/ban", bytes.NewReader(banBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", csrfToken)
	request.AddCookie(cookie)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("expected ban other 403, got %d", response.Code)
	}

	connection.BelfastInstance = connection.NewServer("127.0.0.1", 0, func(*[]byte, *connection.Client, int) {})
	var sink net.Conn = writeSinkConn{}
	client := &connection.Client{Commander: &orm.Commander{CommanderID: 111}, Server: connection.BelfastInstance, Hash: 9001, Connection: &sink}
	connection.BelfastInstance.AddClient(client)

	request = httptest.NewRequest(http.MethodPost, "/api/v1/players/111/kick", nil)
	request.Header.Set("X-CSRF-Token", csrfToken)
	request.AddCookie(cookie)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected kick self 200, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodPost, "/api/v1/players/222/kick", nil)
	request.Header.Set("X-CSRF-Token", csrfToken)
	request.AddCookie(cookie)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("expected kick other 403, got %d", response.Code)
	}
}

type writeSinkConn struct{}

func (writeSinkConn) Read([]byte) (int, error)         { return 0, io.EOF }
func (writeSinkConn) Write(p []byte) (int, error)      { return len(p), nil }
func (writeSinkConn) Close() error                     { return nil }
func (writeSinkConn) LocalAddr() net.Addr              { return dummyAddr("local") }
func (writeSinkConn) RemoteAddr() net.Addr             { return dummyAddr("remote") }
func (writeSinkConn) SetDeadline(time.Time) error      { return nil }
func (writeSinkConn) SetReadDeadline(time.Time) error  { return nil }
func (writeSinkConn) SetWriteDeadline(time.Time) error { return nil }

type dummyAddr string

func (a dummyAddr) Network() string { return string(a) }
func (a dummyAddr) String() string  { return string(a) }

func TestPlayersAPIWithoutPermPlayersReturnsForbidden(t *testing.T) {
	app := newPlayerAuthzTestApp(t)
	clearPlayerAuthzTables(t)
	if err := orm.EnsureAuthzDefaults(); err != nil {
		t.Fatalf("ensure authz defaults: %v", err)
	}

	createPlayerAuthzCommander(t, 111, "Self")
	password := "this-is-a-strong-pass"
	createPlayerAccount(t, 111, password)
	if err := orm.ReplaceRolePolicyByName(authz.RolePlayer, map[string]authz.Capability{}, nil); err != nil {
		t.Fatalf("update policy: %v", err)
	}
	cookie := loginPlayer(t, app, 111, password)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players/111", nil)
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", response.Code)
	}
}

func TestPlayersAPIReadAnySeesNotFoundForMissingCommander(t *testing.T) {
	app := newPlayerAuthzTestApp(t)
	clearPlayerAuthzTables(t)
	if err := orm.EnsureAuthzDefaults(); err != nil {
		t.Fatalf("ensure authz defaults: %v", err)
	}

	createPlayerAuthzCommander(t, 111, "Self")
	password := "this-is-a-strong-pass"
	createPlayerAccount(t, 111, password)
	if err := orm.ReplaceRolePolicyByName(authz.RolePlayer, map[string]authz.Capability{authz.PermPlayers: {ReadAny: true}}, nil); err != nil {
		t.Fatalf("update policy: %v", err)
	}
	cookie := loginPlayer(t, app, 111, password)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players/9999", nil)
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", response.Code)
	}
}
