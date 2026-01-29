package orm

import (
	"encoding/json"
	"testing"
)

func TestJuustagramJSONHelpers(t *testing.T) {
	list := JuustagramUint32List{1, 2}
	value, err := list.Value()
	if err != nil {
		t.Fatalf("value: %v", err)
	}
	var decoded JuustagramUint32List
	if err := decoded.Scan(value); err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(decoded) != 2 {
		t.Fatalf("expected decoded list")
	}
	if err := decoded.Scan(1); err == nil {
		t.Fatalf("expected scan error")
	}
	if err := decoded.Scan(nil); err != nil {
		t.Fatalf("scan nil: %v", err)
	}

	var replies JuustagramReplyList
	if err := replies.UnmarshalJSON([]byte("\"\"")); err != nil {
		t.Fatalf("unmarshal empty: %v", err)
	}
	if len(replies) != 0 {
		t.Fatalf("expected empty replies")
	}
	if err := replies.UnmarshalJSON([]byte("[1,2]")); err != nil {
		t.Fatalf("unmarshal replies: %v", err)
	}
	if len(replies) != 2 {
		t.Fatalf("expected replies length")
	}
	if _, err := replies.Value(); err != nil {
		t.Fatalf("reply list value: %v", err)
	}
	var decodedReplies JuustagramReplyList
	if err := decodedReplies.Scan([]byte("[3]")); err != nil {
		t.Fatalf("reply list scan: %v", err)
	}

	config := JuustagramTimeConfig{{1, 2}}
	if _, err := config.Value(); err != nil {
		t.Fatalf("config value: %v", err)
	}
	var decodedConfig JuustagramTimeConfig
	if err := decodedConfig.Scan([]byte("[[3,4]]")); err != nil {
		t.Fatalf("config scan: %v", err)
	}
}

func TestJuustagramTemplatesAndLanguages(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &JuustagramTemplate{})
	clearTable(t, &JuustagramNpcTemplate{})
	clearTable(t, &JuustagramShipGroupTemplate{})
	clearTable(t, &JuustagramLanguage{})

	template := JuustagramTemplate{ID: 1, GroupID: 1, ShipGroup: 2, Name: "Temp", Sculpture: "S", PicturePersist: "P", MessagePersist: "M", IsActive: 1}
	if err := GormDB.Create(&template).Error; err != nil {
		t.Fatalf("seed template: %v", err)
	}
	if _, err := GetJuustagramTemplate(1); err != nil {
		t.Fatalf("get template: %v", err)
	}
	list, total, err := ListJuustagramTemplates(0, 10)
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatalf("list templates: %v", err)
	}

	npc := JuustagramNpcTemplate{ID: 1, ShipGroup: 2, MessagePersist: "op_reply_5_1", NpcReplyPersist: JuustagramReplyList{1}}
	if err := GormDB.Create(&npc).Error; err != nil {
		t.Fatalf("seed npc template: %v", err)
	}
	if _, err := GetJuustagramNpcTemplate(1); err != nil {
		t.Fatalf("get npc template: %v", err)
	}
	listNpc, totalNpc, err := ListJuustagramNpcTemplates(0, 10)
	if err != nil || totalNpc != 1 || len(listNpc) != 1 {
		t.Fatalf("list npc templates: %v", err)
	}
	replies, err := ListJuustagramOpReplies(5)
	if err != nil || len(replies) != 1 {
		t.Fatalf("list op replies: %v", err)
	}

	group := JuustagramShipGroupTemplate{ShipGroup: 2, Name: "G", Background: "B", Sculpture: "S", SculptureII: "S2", Nationality: 1, Type: 1}
	if err := GormDB.Create(&group).Error; err != nil {
		t.Fatalf("seed ship group: %v", err)
	}
	if _, err := GetJuustagramShipGroupTemplate(2); err != nil {
		t.Fatalf("get ship group: %v", err)
	}
	listGroup, totalGroup, err := ListJuustagramShipGroupTemplates(0, 10)
	if err != nil || totalGroup != 1 || len(listGroup) != 1 {
		t.Fatalf("list ship group templates: %v", err)
	}

	lang := JuustagramLanguage{Key: "juu_test", Value: "Hello"}
	if err := GormDB.Create(&lang).Error; err != nil {
		t.Fatalf("seed language: %v", err)
	}
	value, err := GetJuustagramLanguage("juu_test")
	if err != nil || value != "Hello" {
		t.Fatalf("get language: %v", err)
	}
	langs, err := ListJuustagramLanguageByPrefix("juu_")
	if err != nil || len(langs) != 1 {
		t.Fatalf("list language by prefix: %v", err)
	}
}

func TestJuustagramMessageStateAndDiscuss(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &JuustagramMessageState{})
	clearTable(t, &JuustagramPlayerDiscuss{})

	state, err := GetOrCreateJuustagramMessageState(1, 5, 100)
	if err != nil {
		t.Fatalf("get or create message state: %v", err)
	}
	state.IsRead = 1
	if err := SaveJuustagramMessageState(state); err != nil {
		t.Fatalf("save message state: %v", err)
	}
	if _, err := GetJuustagramMessageState(1, 5); err != nil {
		t.Fatalf("get message state: %v", err)
	}

	entry := JuustagramPlayerDiscuss{CommanderID: 1, MessageID: 5, DiscussID: 1, OptionIndex: 2, NpcReplyID: 3, CommentTime: 100}
	if err := UpsertJuustagramPlayerDiscuss(&entry); err != nil {
		t.Fatalf("upsert discuss: %v", err)
	}
	if _, err := GetJuustagramPlayerDiscuss(1, 5, 1); err != nil {
		t.Fatalf("get discuss: %v", err)
	}
	list, err := ListJuustagramPlayerDiscuss(1, 5)
	if err != nil || len(list) != 1 {
		t.Fatalf("list discuss: %v", err)
	}
}

func TestJuustagramReplyListJSON(t *testing.T) {
	var replies JuustagramReplyList
	data, err := json.Marshal([]uint32{9})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := replies.UnmarshalJSON(data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
}
