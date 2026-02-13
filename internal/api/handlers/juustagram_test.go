package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/orm"
)

var juustagramHandlerTestOnce sync.Once

func initJuustagramHandlerTestDB(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	juustagramHandlerTestOnce.Do(func() {
		orm.InitDatabase()
	})
}

func newJuustagramHandlerTestApp(t *testing.T) *iris.Application {
	initJuustagramHandlerTestDB(t)
	app := iris.New()
	handler := NewJuustagramHandler()
	RegisterJuustagramRoutes(app.Party("/api/v1/juustagram"), handler)
	RegisterJuustagramPlayerRoutes(app.Party("/api/v1/players/{id:uint}/juustagram"), handler)
	if err := app.Build(); err != nil {
		t.Fatalf("build app: %v", err)
	}
	return app
}

func seedJuustagramHandlerData(t *testing.T) {
	t.Helper()
	execTestSQL(t, "DELETE FROM juustagram_player_discusses")
	execTestSQL(t, "DELETE FROM juustagram_message_states")
	execTestSQL(t, "DELETE FROM juustagram_templates")
	execTestSQL(t, "DELETE FROM juustagram_npc_templates")
	execTestSQL(t, "DELETE FROM juustagram_languages")
	execTestSQL(t, "DELETE FROM commanders")

	if err := orm.CreateCommanderRoot(8100, 1, "Juustagram Tester", 0, 0); err != nil {
		t.Fatalf("create commander: %v", err)
	}

	message := orm.JuustagramTemplate{
		ID:                1,
		GroupID:           1,
		ShipGroup:         100100,
		Name:              "Tester",
		Sculpture:         "test",
		PicturePersist:    "https://example.com/test.png",
		MessagePersist:    "ins_1",
		IsActive:          0,
		NpcDiscussPersist: orm.JuustagramUint32List{1},
		TimePersist:       orm.JuustagramTimeConfig{{2024, 1, 1}, {0, 0, 0}},
	}
	if err := orm.CreateJuustagramTemplate(&message); err != nil {
		t.Fatalf("create template: %v", err)
	}

	npcDiscuss := orm.JuustagramNpcTemplate{
		ID:              1,
		ShipGroup:       100100,
		MessagePersist:  "ins_discuss_1_1",
		NpcReplyPersist: orm.JuustagramReplyList{2},
		TimePersist:     orm.JuustagramTimeConfig{{2024, 1, 1}, {1, 0, 0}},
	}
	npcReply := orm.JuustagramNpcTemplate{
		ID:              2,
		ShipGroup:       100100,
		MessagePersist:  "ins_reply_1_1_1",
		NpcReplyPersist: orm.JuustagramReplyList{},
		TimePersist:     orm.JuustagramTimeConfig{{2024, 1, 1}, {1, 5, 0}},
	}
	opReply := orm.JuustagramNpcTemplate{
		ID:              3,
		ShipGroup:       100100,
		MessagePersist:  "op_reply_1_1_1",
		NpcReplyPersist: orm.JuustagramReplyList{},
		TimePersist:     orm.JuustagramTimeConfig{{2024, 1, 1}, {1, 10, 0}},
	}
	if err := orm.CreateJuustagramNpcTemplate(&npcDiscuss); err != nil {
		t.Fatalf("create npc discuss: %v", err)
	}
	if err := orm.CreateJuustagramNpcTemplate(&npcReply); err != nil {
		t.Fatalf("create npc reply: %v", err)
	}
	if err := orm.CreateJuustagramNpcTemplate(&opReply); err != nil {
		t.Fatalf("create op reply: %v", err)
	}

	languageEntries := []orm.JuustagramLanguage{
		{Key: "ins_1", Value: "hello world"},
		{Key: "ins_discuss_1_1", Value: "npc discuss"},
		{Key: "ins_reply_1_1_1", Value: "npc reply"},
		{Key: "op_reply_1_1_1", Value: "op reply"},
		{Key: "ins_op_1_1_1", Value: "player option"},
	}
	for _, entry := range languageEntries {
		if err := orm.CreateJuustagramLanguage(&entry); err != nil {
			t.Fatalf("create language entry: %v", err)
		}
	}
}

func TestJuustagramMessageEndpoints(t *testing.T) {
	app := newJuustagramHandlerTestApp(t)
	seedJuustagramHandlerData(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players/8100/juustagram/messages", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	var listResponse struct {
		OK   bool `json:"ok"`
		Data struct {
			Messages []types.JuustagramMessage `json:"messages"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&listResponse); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(listResponse.Data.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(listResponse.Data.Messages))
	}
	if listResponse.Data.Messages[0].Text != "hello world" {
		t.Fatalf("unexpected message text")
	}

	updatePayload := strings.NewReader("{\"read\": true, \"like\": true}")
	request = httptest.NewRequest(http.MethodPatch, "/api/v1/players/8100/juustagram/messages/1", updatePayload)
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	var updateResponse struct {
		OK   bool `json:"ok"`
		Data struct {
			Message types.JuustagramMessage `json:"message"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&updateResponse); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if updateResponse.Data.Message.IsRead != 1 || updateResponse.Data.Message.IsGood != 1 {
		t.Fatalf("expected read and like flags set")
	}
}
