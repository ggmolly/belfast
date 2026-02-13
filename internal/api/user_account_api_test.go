package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/middleware"
	"github.com/ggmolly/belfast/internal/api/routes"
	"github.com/ggmolly/belfast/internal/auth"
	"github.com/ggmolly/belfast/internal/authz"
	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
)

func newUserTestApp(t *testing.T) *iris.Application {
	t.Helper()
	initAuthTestDB(t)
	app := iris.New()
	cfg := &config.Config{
		Auth: config.AuthConfig{
			CookieName: "belfast_session",
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
		"audit_logs",
		"sessions",
		"user_registration_challenges",
		"account_permission_overrides",
		"account_roles",
		"accounts",
		"mail_attachments",
		"mails",
		"commanders",
		"owned_resources",
		"resources",
		"commander_items",
		"items",
	}
	for _, table := range tables {
		execAPITestSQLT(t, fmt.Sprintf("DELETE FROM %s", table))
	}
	execAPITestSQLT(t, "DELETE FROM role_permissions WHERE role_id = (SELECT id FROM roles WHERE name = $1)", authz.RolePlayer)
}

func createUserTestCommander(t *testing.T, commanderID uint32, name string) {
	t.Helper()
	if err := orm.CreateCommanderRoot(commanderID, commanderID, name, 0, 0); err != nil {
		t.Fatalf("create commander root: %v", err)
	}
	execAPITestSQLT(
		t,
		"UPDATE commanders SET level = $1, display_icon_id = $2, display_skin_id = $3, selected_icon_frame_id = $4, selected_chat_frame_id = $5, display_icon_theme_id = $6 WHERE commander_id = $7",
		int64(10),
		int64(1001),
		int64(1001),
		int64(200),
		int64(300),
		int64(400),
		int64(commanderID),
	)
}

func TestUserRegistrationChallengeAndStatus(t *testing.T) {
	app := newUserTestApp(t)
	clearUserTables(t)

	createUserTestCommander(t, 9001, "Registration User")

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
			ExpiresAt   string `json:"expires_at"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&createResponse); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if createResponse.Data.ChallengeID == "" {
		t.Fatalf("expected challenge id")
	}
	if createResponse.Data.ExpiresAt == "" {
		t.Fatalf("expected expires_at")
	}

	var challengePin string
	if err := db.DefaultStore.Pool.QueryRow(context.Background(), "SELECT pin FROM user_registration_challenges WHERE id = $1", createResponse.Data.ChallengeID).Scan(&challengePin); err != nil {
		t.Fatalf("load challenge: %v", err)
	}
	var mailBody string
	if err := db.DefaultStore.Pool.QueryRow(context.Background(), "SELECT body FROM mails WHERE receiver_id = $1 ORDER BY id LIMIT 1", int64(9001)).Scan(&mailBody); err != nil {
		t.Fatalf("load mail: %v", err)
	}
	if !strings.Contains(mailBody, challengePin) {
		t.Fatalf("expected mail body to include pin")
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

func TestUserRegistrationVerifyChallenge(t *testing.T) {
	app := newUserTestApp(t)
	clearUserTables(t)

	createUserTestCommander(t, 9002, "Verify User")

	payload := []byte(`{"commander_id":9002,"password":"this-is-a-strong-pass"}`)
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
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&createResponse); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if createResponse.Data.ChallengeID == "" {
		t.Fatalf("expected challenge id")
	}

	var challengeID string
	var challengePin string
	if err := db.DefaultStore.Pool.QueryRow(context.Background(), "SELECT id, pin FROM user_registration_challenges WHERE id = $1", createResponse.Data.ChallengeID).Scan(&challengeID, &challengePin); err != nil {
		t.Fatalf("load challenge: %v", err)
	}

	verifyPayload := []byte(`{"pin":"000000"}`)
	request = httptest.NewRequest(http.MethodPost, "/api/v1/registration/challenges/"+createResponse.Data.ChallengeID+"/verify", bytes.NewReader(verifyPayload))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected verify 400, got %d", response.Code)
	}

	verifyPayload = []byte("{\"pin\":\"B-" + challengePin + "\"}")
	request = httptest.NewRequest(http.MethodPost, "/api/v1/registration/challenges/"+createResponse.Data.ChallengeID+"/verify", bytes.NewReader(verifyPayload))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected verify 200, got %d", response.Code)
	}
	var verifyResponse struct {
		OK   bool `json:"ok"`
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&verifyResponse); err != nil {
		t.Fatalf("decode verify response: %v", err)
	}
	if verifyResponse.Data.Status != orm.UserRegistrationStatusConsumed {
		t.Fatalf("expected consumed status, got %s", verifyResponse.Data.Status)
	}

	accountCount := queryAPITestInt64(t, "SELECT COUNT(*) FROM accounts WHERE commander_id = $1", int64(9002))
	if accountCount == 0 {
		t.Fatalf("expected user account created")
	}
	var challengeStatus string
	if err := db.DefaultStore.Pool.QueryRow(context.Background(), "SELECT status FROM user_registration_challenges WHERE id = $1", challengeID).Scan(&challengeStatus); err != nil {
		t.Fatalf("reload challenge: %v", err)
	}
	if challengeStatus != orm.UserRegistrationStatusConsumed {
		t.Fatalf("expected challenge consumed, got %s", challengeStatus)
	}
}

func TestUserRegistrationChallengeReissue(t *testing.T) {
	app := newUserTestApp(t)
	clearUserTables(t)

	createUserTestCommander(t, 9003, "Reissue User")

	payload := []byte(`{"commander_id":9003,"password":"this-is-a-strong-pass"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/registration/challenges", bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected create challenge 200, got %d", response.Code)
	}
	var firstResponse struct {
		OK   bool `json:"ok"`
		Data struct {
			ChallengeID string `json:"challenge_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&firstResponse); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if firstResponse.Data.ChallengeID == "" {
		t.Fatalf("expected challenge id")
	}

	request = httptest.NewRequest(http.MethodPost, "/api/v1/registration/challenges", bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected reissue 200, got %d", response.Code)
	}
	var secondResponse struct {
		OK   bool `json:"ok"`
		Data struct {
			ChallengeID string `json:"challenge_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&secondResponse); err != nil {
		t.Fatalf("decode reissue response: %v", err)
	}
	if secondResponse.Data.ChallengeID == "" {
		t.Fatalf("expected reissue challenge id")
	}
	if secondResponse.Data.ChallengeID == firstResponse.Data.ChallengeID {
		t.Fatalf("expected new challenge id")
	}

	var firstChallengeStatus string
	if err := db.DefaultStore.Pool.QueryRow(context.Background(), "SELECT status FROM user_registration_challenges WHERE id = $1", firstResponse.Data.ChallengeID).Scan(&firstChallengeStatus); err != nil {
		t.Fatalf("load first challenge: %v", err)
	}
	if firstChallengeStatus != orm.UserRegistrationStatusExpired {
		t.Fatalf("expected first challenge expired, got %s", firstChallengeStatus)
	}
	var secondChallengeStatus string
	if err := db.DefaultStore.Pool.QueryRow(context.Background(), "SELECT status FROM user_registration_challenges WHERE id = $1", secondResponse.Data.ChallengeID).Scan(&secondChallengeStatus); err != nil {
		t.Fatalf("load second challenge: %v", err)
	}
	if secondChallengeStatus != orm.UserRegistrationStatusPending {
		t.Fatalf("expected second challenge pending, got %s", secondChallengeStatus)
	}

	mailCount := queryAPITestInt64(t, "SELECT COUNT(*) FROM mails WHERE receiver_id = $1", int64(9003))
	if mailCount != 2 {
		t.Fatalf("expected 2 mails, got %d", mailCount)
	}
}

func TestUserLoginAndPermissions(t *testing.T) {
	app := newUserTestApp(t)
	clearUserTables(t)
	seedDb()

	createUserTestCommander(t, 9102, "User Login")

	passwordHash, algo, err := auth.HashPassword("this-is-a-strong-pass", auth.NormalizeConfig(config.AuthConfig{}))
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	commanderID := uint32(9102)
	user := orm.Account{
		ID:                uuid.NewString(),
		CommanderID:       &commanderID,
		PasswordHash:      passwordHash,
		PasswordAlgo:      algo,
		PasswordUpdatedAt: time.Now().UTC(),
	}
	if err := orm.CreateAccount(&user); err != nil {
		t.Fatalf("create user account: %v", err)
	}
	if err := orm.AssignRoleByName(user.ID, authz.RolePlayer); err != nil {
		t.Fatalf("assign role: %v", err)
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

	if err := orm.ReplaceRolePolicyByName(authz.RolePlayer, map[string]authz.Capability{authz.PermMeResources: {ReadSelf: true, WriteSelf: true}}, nil); err != nil {
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
