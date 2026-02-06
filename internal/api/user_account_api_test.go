package api_test

import (
	"bytes"
	"encoding/json"
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
		if err := orm.GormDB.Exec("DELETE FROM " + table).Error; err != nil {
			t.Fatalf("clear %s: %v", table, err)
		}
	}
	if err := orm.GormDB.Exec("DELETE FROM role_permissions WHERE role_id = (SELECT id FROM roles WHERE name = ?)", authz.RolePlayer).Error; err != nil {
		t.Fatalf("clear role_permissions: %v", err)
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

	var challenge orm.UserRegistrationChallenge
	if err := orm.GormDB.First(&challenge, "id = ?", createResponse.Data.ChallengeID).Error; err != nil {
		t.Fatalf("load challenge: %v", err)
	}
	var mail orm.Mail
	if err := orm.GormDB.First(&mail, "receiver_id = ?", commander.CommanderID).Error; err != nil {
		t.Fatalf("load mail: %v", err)
	}
	if !strings.Contains(mail.Body, challenge.Pin) {
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

	commander := orm.Commander{
		CommanderID:         9002,
		AccountID:           9002,
		Name:                "Verify User",
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

	var challenge orm.UserRegistrationChallenge
	if err := orm.GormDB.First(&challenge, "id = ?", createResponse.Data.ChallengeID).Error; err != nil {
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

	verifyPayload = []byte("{\"pin\":\"B-" + challenge.Pin + "\"}")
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

	var account orm.Account
	if err := orm.GormDB.First(&account, "commander_id = ?", commander.CommanderID).Error; err != nil {
		t.Fatalf("expected user account created, got %v", err)
	}
	if err := orm.GormDB.First(&challenge, "id = ?", challenge.ID).Error; err != nil {
		t.Fatalf("reload challenge: %v", err)
	}
	if challenge.Status != orm.UserRegistrationStatusConsumed {
		t.Fatalf("expected challenge consumed, got %s", challenge.Status)
	}
}

func TestUserRegistrationChallengeReissue(t *testing.T) {
	app := newUserTestApp(t)
	clearUserTables(t)

	commander := orm.Commander{
		CommanderID:         9003,
		AccountID:           9003,
		Name:                "Reissue User",
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

	var firstChallenge orm.UserRegistrationChallenge
	if err := orm.GormDB.First(&firstChallenge, "id = ?", firstResponse.Data.ChallengeID).Error; err != nil {
		t.Fatalf("load first challenge: %v", err)
	}
	if firstChallenge.Status != orm.UserRegistrationStatusExpired {
		t.Fatalf("expected first challenge expired, got %s", firstChallenge.Status)
	}
	var secondChallenge orm.UserRegistrationChallenge
	if err := orm.GormDB.First(&secondChallenge, "id = ?", secondResponse.Data.ChallengeID).Error; err != nil {
		t.Fatalf("load second challenge: %v", err)
	}
	if secondChallenge.Status != orm.UserRegistrationStatusPending {
		t.Fatalf("expected second challenge pending, got %s", secondChallenge.Status)
	}

	var mailCount int64
	if err := orm.GormDB.Model(&orm.Mail{}).Where("receiver_id = ?", commander.CommanderID).Count(&mailCount).Error; err != nil {
		t.Fatalf("count mails: %v", err)
	}
	if mailCount != 2 {
		t.Fatalf("expected 2 mails, got %d", mailCount)
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

	passwordHash, algo, err := auth.HashPassword("this-is-a-strong-pass", auth.NormalizeConfig(config.AuthConfig{}))
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	commanderID := commander.CommanderID
	user := orm.Account{
		ID:                uuid.NewString(),
		CommanderID:       &commanderID,
		PasswordHash:      passwordHash,
		PasswordAlgo:      algo,
		PasswordUpdatedAt: time.Now().UTC(),
	}
	if err := orm.GormDB.Create(&user).Error; err != nil {
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
