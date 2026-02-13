package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/orm"
)

func seedJuustagramChatData(t *testing.T) uint32 {
	t.Helper()
	execTestSQL(t, "DELETE FROM juustagram_replies")
	execTestSQL(t, "DELETE FROM juustagram_chat_groups")
	execTestSQL(t, "DELETE FROM juustagram_groups")
	execTestSQL(t, "DELETE FROM commanders")
	commanderID := uint32(8200)
	if err := orm.CreateCommanderRoot(commanderID, 1, "Juustagram Chat Tester", 0, 0); err != nil {
		t.Fatalf("create commander: %v", err)
	}
	if _, err := orm.CreateJuustagramGroup(commanderID, 960007, 1); err != nil {
		t.Fatalf("create juustagram group: %v", err)
	}
	if _, err := orm.AddJuustagramChatReply(commanderID, 1, 1, 1, 0); err != nil {
		t.Fatalf("create chat reply: %v", err)
	}
	return commanderID
}

func TestJuustagramChatEndpoints(t *testing.T) {
	app := newJuustagramHandlerTestApp(t)
	commanderID := seedJuustagramChatData(t)
	groups, _, err := orm.ListJuustagramGroups(commanderID, 0, 10)
	if err != nil {
		t.Fatalf("list juustagram groups: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected seeded group, got %d", len(groups))
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players/8200/juustagram/groups?limit=10", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	rawBody := response.Body.Bytes()
	var listResponse struct {
		OK   bool `json:"ok"`
		Data struct {
			Groups []types.JuustagramGroup `json:"groups"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&listResponse); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(listResponse.Data.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d (body: %s)", len(listResponse.Data.Groups), string(rawBody))
	}
	if len(listResponse.Data.Groups[0].ChatGroups) != 1 {
		t.Fatalf("expected 1 chat group")
	}

	payload := strings.NewReader("{\"chat_id\": 2, \"value\": 3}")
	request = httptest.NewRequest(http.MethodPost, "/api/v1/players/8200/juustagram/chat-groups/1/reply", payload)
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	var replyResponse struct {
		OK   bool `json:"ok"`
		Data struct {
			Group types.JuustagramGroup `json:"group"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&replyResponse); err != nil {
		t.Fatalf("decode reply response: %v", err)
	}
	if len(replyResponse.Data.Group.ChatGroups) == 0 || len(replyResponse.Data.Group.ChatGroups[0].ReplyList) != 2 {
		t.Fatalf("expected 2 replies after update")
	}

	readPayload := strings.NewReader("{\"chat_group_ids\": [1]}")
	request = httptest.NewRequest(http.MethodPatch, "/api/v1/players/8200/juustagram/chat-groups/read", readPayload)
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	group, err := orm.GetJuustagramChatGroup(commanderID, 1)
	if err != nil {
		t.Fatalf("load chat group: %v", err)
	}
	if group.ReadFlag != 1 {
		t.Fatalf("expected read flag to be set")
	}
}
