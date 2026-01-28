package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/orm"
)

func newNoticeTestApp(t *testing.T) *iris.Application {
	initPlayerHandlerTestDB(t)
	app := iris.New()
	handler := NewNoticeHandler()
	RegisterNoticeRoutes(app.Party("/api/v1/notices"), handler)
	if err := app.Build(); err != nil {
		t.Fatalf("build app: %v", err)
	}
	return app
}

func clearNotices(t *testing.T) {
	t.Helper()
	if err := orm.GormDB.Exec("DELETE FROM notices").Error; err != nil {
		t.Fatalf("clear notices: %v", err)
	}
}

func seedNotice(t *testing.T, id int, version string, title string, content string) {
	t.Helper()
	notice := orm.Notice{
		ID:       id,
		Version:  version,
		BtnTitle: "Button",
		Title:    title,
		TimeDesc: "2026-01-28",
		Content:  content,
		TagType:  1,
		Icon:     1,
		Track:    "track1",
	}
	if err := orm.GormDB.Create(&notice).Error; err != nil {
		t.Fatalf("seed notice: %v", err)
	}
}

func TestListNoticesReturnsEmpty(t *testing.T) {
	app := newNoticeTestApp(t)
	clearNotices(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/notices", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK   bool `json:"ok"`
		Data struct {
			Notices []struct {
				ID         int    `json:"id"`
				Version    string `json:"version"`
				BtnTitle   string `json:"btn_title"`
				Title      string `json:"title"`
				TitleImage string `json:"title_image"`
				TimeDesc   string `json:"time_desc"`
				Content    string `json:"content"`
				TagType    int    `json:"tag_type"`
				Icon       int    `json:"icon"`
				Track      string `json:"track"`
			} `json:"notices"`
			Meta struct {
				Offset uint32 `json:"offset"`
				Limit  uint32 `json:"limit"`
				Total  int64  `json:"total"`
			} `json:"meta"`
		} `json:"data"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}
	if len(responseStruct.Data.Notices) != 0 {
		t.Fatalf("expected empty notices list, got %d", len(responseStruct.Data.Notices))
	}
	if responseStruct.Data.Meta.Total != 0 {
		t.Fatalf("expected total 0, got %d", responseStruct.Data.Meta.Total)
	}
}

func TestListNoticesReturnsData(t *testing.T) {
	app := newNoticeTestApp(t)
	clearNotices(t)
	seedNotice(t, 1, "1.0", "Test Notice 1", "Content 1")
	seedNotice(t, 2, "1.0", "Test Notice 2", "Content 2")
	t.Cleanup(func() {
		clearNotices(t)
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/notices", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK   bool `json:"ok"`
		Data struct {
			Notices []struct {
				ID         int    `json:"id"`
				Version    string `json:"version"`
				BtnTitle   string `json:"btn_title"`
				Title      string `json:"title"`
				TitleImage string `json:"title_image"`
				TimeDesc   string `json:"time_desc"`
				Content    string `json:"content"`
				TagType    int    `json:"tag_type"`
				Icon       int    `json:"icon"`
				Track      string `json:"track"`
			} `json:"notices"`
			Meta struct {
				Offset uint32 `json:"offset"`
				Limit  uint32 `json:"limit"`
				Total  int64  `json:"total"`
			} `json:"meta"`
		} `json:"data"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}
	if len(responseStruct.Data.Notices) != 2 {
		t.Fatalf("expected 2 notices, got %d", len(responseStruct.Data.Notices))
	}
	if responseStruct.Data.Meta.Total != 2 {
		t.Fatalf("expected total 2, got %d", responseStruct.Data.Meta.Total)
	}
	if responseStruct.Data.Notices[0].ID != 1 && responseStruct.Data.Notices[1].ID != 1 {
		t.Fatalf("expected one of the notices to have id 1")
	}
}

func TestActiveNoticesReturnsEmpty(t *testing.T) {
	app := newNoticeTestApp(t)
	clearNotices(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/notices/active", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK   bool `json:"ok"`
		Data []struct {
			ID         int    `json:"id"`
			Version    string `json:"version"`
			BtnTitle   string `json:"btn_title"`
			Title      string `json:"title"`
			TitleImage string `json:"title_image"`
			TimeDesc   string `json:"time_desc"`
			Content    string `json:"content"`
			TagType    int    `json:"tag_type"`
			Icon       int    `json:"icon"`
			Track      string `json:"track"`
		} `json:"data"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}
	if len(responseStruct.Data) != 0 {
		t.Fatalf("expected empty active notices list, got %d", len(responseStruct.Data))
	}
}

func TestActiveNoticesReturnsData(t *testing.T) {
	app := newNoticeTestApp(t)
	clearNotices(t)
	seedNotice(t, 1, "2.0", "Active Notice 1", "Active Content 1")
	t.Cleanup(func() {
		clearNotices(t)
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/notices/active", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK   bool `json:"ok"`
		Data []struct {
			ID         int    `json:"id"`
			Version    string `json:"version"`
			BtnTitle   string `json:"btn_title"`
			Title      string `json:"title"`
			TitleImage string `json:"title_image"`
			TimeDesc   string `json:"time_desc"`
			Content    string `json:"content"`
			TagType    int    `json:"tag_type"`
			Icon       int    `json:"icon"`
			Track      string `json:"track"`
		} `json:"data"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}
	if len(responseStruct.Data) != 1 {
		t.Fatalf("expected 1 active notice, got %d", len(responseStruct.Data))
	}
	if responseStruct.Data[0].ID != 1 {
		t.Fatalf("expected active notice id 1, got %d", responseStruct.Data[0].ID)
	}
	if responseStruct.Data[0].Version != "2.0" {
		t.Fatalf("expected version '2.0', got %s", responseStruct.Data[0].Version)
	}
}

func TestCreateNotice(t *testing.T) {
	app := newNoticeTestApp(t)
	clearNotices(t)
	t.Cleanup(func() {
		clearNotices(t)
	})

	requestBody := `{"version":"1.0","btn_title":"Test Button","title":"Test Title","content":"Test Content","tag_type":1,"icon":1}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/notices", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK bool `json:"ok"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}

	var notice orm.Notice
	if err := orm.GormDB.Where("title = ?", "Test Title").First(&notice).Error; err != nil {
		t.Fatalf("query notice failed: %v", err)
	}
	if notice.Title != "Test Title" {
		t.Fatalf("expected title 'Test Title', got %s", notice.Title)
	}
	if notice.Content != "Test Content" {
		t.Fatalf("expected content 'Test Content', got %s", notice.Content)
	}
	if notice.TagType != 1 {
		t.Fatalf("expected tag_type 1, got %d", notice.TagType)
	}
}

func TestCreateNoticeMissingTitle(t *testing.T) {
	app := newNoticeTestApp(t)
	clearNotices(t)

	requestBody := `{"version":"1.0","btn_title":"Test Button","content":"Test Content","tag_type":1,"icon":1}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/notices", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}

	var responseStruct struct {
		OK bool `json:"ok"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if responseStruct.OK {
		t.Fatalf("expected ok false for missing title")
	}
}

func TestCreateNoticeMissingContent(t *testing.T) {
	app := newNoticeTestApp(t)
	clearNotices(t)

	requestBody := `{"version":"1.0","btn_title":"Test Button","title":"Test Title","tag_type":1,"icon":1}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/notices", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}

	var responseStruct struct {
		OK bool `json:"ok"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if responseStruct.OK {
		t.Fatalf("expected ok false for missing content")
	}
}

func TestUpdateNotice(t *testing.T) {
	app := newNoticeTestApp(t)
	clearNotices(t)
	seedNotice(t, 1, "1.0", "Old Title", "Old Content")
	t.Cleanup(func() {
		clearNotices(t)
	})

	requestBody := `{"version":"2.0","btn_title":"New Button","title":"New Title","content":"New Content","tag_type":2,"icon":2}`
	request := httptest.NewRequest(http.MethodPut, "/api/v1/notices/1", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK bool `json:"ok"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}

	var notice orm.Notice
	if err := orm.GormDB.First(&notice, 1).Error; err != nil {
		t.Fatalf("query notice failed: %v", err)
	}
	if notice.Title != "New Title" {
		t.Fatalf("expected title 'New Title' after update, got %s", notice.Title)
	}
	if notice.Content != "New Content" {
		t.Fatalf("expected content 'New Content' after update, got %s", notice.Content)
	}
	if notice.TagType != 2 {
		t.Fatalf("expected tag_type 2 after update, got %d", notice.TagType)
	}
}

func TestUpdateNoticeNotFound(t *testing.T) {
	app := newNoticeTestApp(t)
	clearNotices(t)

	requestBody := `{"version":"2.0","btn_title":"New Button","title":"New Title","content":"New Content","tag_type":2,"icon":2}`
	request := httptest.NewRequest(http.MethodPut, "/api/v1/notices/999", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", response.Code)
	}

	var responseStruct struct {
		OK bool `json:"ok"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if responseStruct.OK {
		t.Fatalf("expected ok false for not found")
	}
}

func TestDeleteNotice(t *testing.T) {
	app := newNoticeTestApp(t)
	clearNotices(t)
	seedNotice(t, 1, "1.0", "Delete Me", "Content to delete")
	t.Cleanup(func() {
		clearNotices(t)
	})

	request := httptest.NewRequest(http.MethodDelete, "/api/v1/notices/1", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var responseStruct struct {
		OK bool `json:"ok"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true")
	}

	var notice orm.Notice
	if err := orm.GormDB.First(&notice, 1).Error; err == nil {
		t.Fatalf("expected notice to be deleted")
	}
}

func TestDeleteNoticeNotFound(t *testing.T) {
	app := newNoticeTestApp(t)
	clearNotices(t)

	request := httptest.NewRequest(http.MethodDelete, "/api/v1/notices/999", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200 (delete is idempotent), got %d", response.Code)
	}

	var responseStruct struct {
		OK bool `json:"ok"`
	}

	if err := json.NewDecoder(response.Body).Decode(&responseStruct); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if !responseStruct.OK {
		t.Fatalf("expected ok true (delete is idempotent)")
	}
}
