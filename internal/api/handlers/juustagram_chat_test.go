package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/orm"
)

func seedJuustagramChatData(t *testing.T) uint32 {
	t.Helper()
	if err := orm.GormDB.Exec("DELETE FROM juustagram_replies").Error; err != nil {
		t.Fatalf("clear juustagram replies: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM juustagram_chat_groups").Error; err != nil {
		t.Fatalf("clear juustagram chat groups: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM juustagram_groups").Error; err != nil {
		t.Fatalf("clear juustagram groups: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM commanders").Error; err != nil {
		t.Fatalf("clear commanders: %v", err)
	}
	commanderID := uint32(8200)
	commander := orm.Commander{
		CommanderID: commanderID,
		AccountID:   1,
		Level:       1,
		Exp:         0,
		Name:        "Juustagram Chat Tester",
		LastLogin:   time.Now().UTC(),
	}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}
	group := orm.JuustagramGroup{
		CommanderID:  commanderID,
		GroupID:      960007,
		SkinID:       0,
		Favorite:     0,
		CurChatGroup: 1,
	}
	if err := orm.GormDB.Create(&group).Error; err != nil {
		t.Fatalf("create juustagram group: %v", err)
	}
	chatGroup := orm.JuustagramChatGroup{
		CommanderID:   commanderID,
		GroupRecordID: group.ID,
		ChatGroupID:   1,
		OpTime:        0,
		ReadFlag:      0,
	}
	if err := orm.GormDB.Create(&chatGroup).Error; err != nil {
		t.Fatalf("create juustagram chat group: %v", err)
	}
	reply := orm.JuustagramReply{
		ChatGroupRecordID: chatGroup.ID,
		Sequence:          1,
		Key:               1,
		Value:             1,
	}
	if err := orm.GormDB.Create(&reply).Error; err != nil {
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
