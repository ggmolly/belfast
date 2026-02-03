package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/middleware"
	"github.com/ggmolly/belfast/internal/api/routes"
	"github.com/ggmolly/belfast/internal/auth"
	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/orm"
)

func newUserTestApp(t *testing.T) *iris.Application {
	t.Helper()
	initAuthTestDB(t)
	app := iris.New()
	cfg := &config.Config{
		Auth: config.AuthConfig{
			CookieName: "belfast_admin_session",
		},
		UserAuth: config.AuthConfig{
			CookieName: "belfast_user_session",
		},
	}
	app.UseRouter(middleware.Auth(cfg))
	routes.RegisterRegistration(app, cfg)
	routes.RegisterUserAuth(app, cfg)
	routes.RegisterMe(app, cfg)
	if err := app.Build(); err != nil {
		t.Fatalf("build app: %v", err)
	}
	return app
}

func clearUserTables(t *testing.T) {
	t.Helper()
	tables := []string{
		"user_audit_logs",
		"user_sessions",
		"user_registration_challenges",
		"user_permission_policies",
		"user_accounts",
		"commanders",
		"owned_resources",
		"resources",
		"commander_items",
		"items",
	}
	for _, table := range tables {
		if err := orm.GormDB.Exec("DELETE FROM " + table).Error; err != nil {
			t.Fatalf("clear %s: %v", table, err)
		}
	}
}

func TestUserRegistrationChallengeAndStatus(t *testing.T) {
	app := newUserTestApp(t)
	clearUserTables(t)

	commander := orm.Commander{
		CommanderID:         9001,
		AccountID:           9001,
		Name:                "Registration User",
		Level:               10,
		DisplayIconID:       1001,
		DisplaySkinID:       1001,
		SelectedIconFrameID: 200,
		SelectedChatFrameID: 300,
		DisplayIconThemeID:  400,
	}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}

	payload := []byte(`{"commander_id":9001,"password":"this-is-a-strong-pass"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/registration/challenges", bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected create challenge 200, got %d", response.Code)
	}
	var createResponse struct {
		OK   bool `json:"ok"`
		Data struct {
			ChallengeID string `json:"challenge_id"`
			Pin         string `json:"pin"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&createResponse); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if createResponse.Data.ChallengeID == "" {
		t.Fatalf("expected challenge id")
	}
	if len(createResponse.Data.Pin) != 8 || createResponse.Data.Pin[:2] != "B-" {
		t.Fatalf("expected pin with B- prefix, got %q", createResponse.Data.Pin)
	}

	statusRequest := httptest.NewRequest(http.MethodGet, "/api/v1/registration/challenges/"+createResponse.Data.ChallengeID, nil)
	statusResponse := httptest.NewRecorder()
	app.ServeHTTP(statusResponse, statusRequest)
	if statusResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", statusResponse.Code)
	}
	var statusPayload struct {
		OK   bool `json:"ok"`
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.NewDecoder(statusResponse.Body).Decode(&statusPayload); err != nil {
		t.Fatalf("decode status response: %v", err)
	}
	if statusPayload.Data.Status != orm.UserRegistrationStatusPending {
		t.Fatalf("expected pending status, got %s", statusPayload.Data.Status)
	}
}

func TestUserLoginAndPermissions(t *testing.T) {
	app := newUserTestApp(t)
	clearUserTables(t)
	seedDb()

	commander := orm.Commander{
		CommanderID:         9102,
		AccountID:           9102,
		Name:                "User Login",
		Level:               10,
		DisplayIconID:       1001,
		DisplaySkinID:       1001,
		SelectedIconFrameID: 200,
		SelectedChatFrameID: 300,
		DisplayIconThemeID:  400,
	}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}

	passwordHash, algo, err := auth.HashPassword("this-is-a-strong-pass", auth.NormalizeUserConfig(config.AuthConfig{}))
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user := orm.UserAccount{
		ID:                uuid.NewString(),
		CommanderID:       commander.CommanderID,
		PasswordHash:      passwordHash,
		PasswordAlgo:      algo,
		PasswordUpdatedAt: time.Now().UTC(),
	}
	if err := orm.GormDB.Create(&user).Error; err != nil {
		t.Fatalf("create user account: %v", err)
	}

	loginPayload := []byte(`{"commander_id":9102,"password":"this-is-a-strong-pass"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/user/auth/login", bytes.NewReader(loginPayload))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d", response.Code)
	}
	cookies := response.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected user session cookie")
	}

	resourceRequest := httptest.NewRequest(http.MethodGet, "/api/v1/me/resources", nil)
	resourceRequest.AddCookie(cookies[0])
	resourceResponse := httptest.NewRecorder()
	app.ServeHTTP(resourceResponse, resourceRequest)
	if resourceResponse.Code != http.StatusForbidden {
		t.Fatalf("expected resources 403, got %d", resourceResponse.Code)
	}

	if _, err := orm.UpdateUserPermissionPolicy([]string{"self.resources.read", "self.resources.update"}, nil); err != nil {
		t.Fatalf("update permission policy: %v", err)
	}

	resourceRequest = httptest.NewRequest(http.MethodGet, "/api/v1/me/resources", nil)
	resourceRequest.AddCookie(cookies[0])
	resourceResponse = httptest.NewRecorder()
	app.ServeHTTP(resourceResponse, resourceRequest)
	if resourceResponse.Code != http.StatusOK {
		t.Fatalf("expected resources 200, got %d", resourceResponse.Code)
	}

	sessionRequest := httptest.NewRequest(http.MethodGet, "/api/v1/user/auth/session", nil)
	sessionRequest.AddCookie(cookies[0])
	sessionResponse := httptest.NewRecorder()
	app.ServeHTTP(sessionResponse, sessionRequest)
	if sessionResponse.Code != http.StatusOK {
		t.Fatalf("expected session 200, got %d", sessionResponse.Code)
	}
	var sessionPayload struct {
		OK   bool `json:"ok"`
		Data struct {
			CSRFToken string `json:"csrf_token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(sessionResponse.Body).Decode(&sessionPayload); err != nil {
		t.Fatalf("decode session response: %v", err)
	}
	if sessionPayload.Data.CSRFToken == "" {
		t.Fatalf("expected csrf token")
	}

	updatePayload := []byte(`{"resources":[{"resource_id":1,"amount":999}]}`)
	updateRequest := httptest.NewRequest(http.MethodPut, "/api/v1/me/resources", bytes.NewReader(updatePayload))
	updateRequest.Header.Set("Content-Type", "application/json")
	updateRequest.Header.Set("X-CSRF-Token", sessionPayload.Data.CSRFToken)
	updateRequest.AddCookie(cookies[0])
	updateResponse := httptest.NewRecorder()
	app.ServeHTTP(updateResponse, updateRequest)
	if updateResponse.Code != http.StatusOK {
		t.Fatalf("expected update 200, got %d", updateResponse.Code)
	}
}
