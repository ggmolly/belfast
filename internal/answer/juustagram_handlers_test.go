package answer

import (
	"sync"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

var juustagramHandlerTestOnce sync.Once

func initJuustagramHandlerTestDB(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	juustagramHandlerTestOnce.Do(func() {
		orm.InitDatabase()
	})
}

func seedJuustagramHandlerData(t *testing.T) {
	t.Helper()
	if err := orm.GormDB.Exec("DELETE FROM juustagram_player_discusses").Error; err != nil {
		t.Fatalf("clear juustagram player discuss: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM juustagram_message_states").Error; err != nil {
		t.Fatalf("clear juustagram message state: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM juustagram_templates").Error; err != nil {
		t.Fatalf("clear juustagram templates: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM juustagram_npc_templates").Error; err != nil {
		t.Fatalf("clear juustagram npc templates: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM juustagram_languages").Error; err != nil {
		t.Fatalf("clear juustagram language: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM commanders").Error; err != nil {
		t.Fatalf("clear commanders: %v", err)
	}

	commander := orm.Commander{
		CommanderID: 1001,
		AccountID:   1,
		Level:       1,
		Exp:         0,
		Name:        "Tester",
		LastLogin:   time.Now().UTC(),
	}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
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
	if err := orm.GormDB.Create(&message).Error; err != nil {
		t.Fatalf("create juustagram template: %v", err)
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
	if err := orm.GormDB.Create(&npcDiscuss).Error; err != nil {
		t.Fatalf("create npc discuss: %v", err)
	}
	if err := orm.GormDB.Create(&npcReply).Error; err != nil {
		t.Fatalf("create npc reply: %v", err)
	}
	if err := orm.GormDB.Create(&opReply).Error; err != nil {
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
		if err := orm.GormDB.Create(&entry).Error; err != nil {
			t.Fatalf("create language entry: %v", err)
		}
	}
}

func TestJuustagramMessageRange(t *testing.T) {
	initJuustagramHandlerTestDB(t)
	seedJuustagramHandlerData(t)
	commander := &orm.Commander{CommanderID: 1001}
	client := &connection.Client{Commander: commander}
	payload := protobuf.CS_11705{
		IndexBegin: proto.Uint32(1),
		IndexEnd:   proto.Uint32(1),
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	_, packetID, err := JuustagramMessageRange(&buffer, client)
	if err != nil {
		t.Fatalf("juustagram message range failed: %v", err)
	}
	if packetID != consts.JuustagramPacketRangeResp {
		t.Fatalf("expected packet %d, got %d", consts.JuustagramPacketRangeResp, packetID)
	}
	data := client.Buffer.Bytes()
	if len(data) < 7 {
		t.Fatalf("expected response payload")
	}
	data = data[7:]
	var response protobuf.SC_11706
	if err := proto.Unmarshal(data, &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(response.GetInsMessageList()) != 1 {
		t.Fatalf("expected 1 message, got %d", len(response.GetInsMessageList()))
	}
	if response.GetInsMessageList()[0].GetText() != "hello world" {
		t.Fatalf("unexpected message text")
	}
}

func TestJuustagramMessageRangeSkipsUnpublished(t *testing.T) {
	initJuustagramHandlerTestDB(t)
	if err := orm.GormDB.Exec("DELETE FROM juustagram_templates").Error; err != nil {
		t.Fatalf("clear juustagram templates: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM juustagram_languages").Error; err != nil {
		t.Fatalf("clear juustagram language: %v", err)
	}
	message := orm.JuustagramTemplate{
		ID:             620,
		GroupID:        620,
		MessagePersist: "",
	}
	if err := orm.GormDB.Create(&message).Error; err != nil {
		t.Fatalf("create juustagram template: %v", err)
	}

	commander := &orm.Commander{CommanderID: 1001}
	client := &connection.Client{Commander: commander}
	payload := protobuf.CS_11705{
		IndexBegin: proto.Uint32(620),
		IndexEnd:   proto.Uint32(620),
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	_, packetID, err := JuustagramMessageRange(&buffer, client)
	if err != nil {
		t.Fatalf("juustagram message range failed: %v", err)
	}
	if packetID != consts.JuustagramPacketRangeResp {
		t.Fatalf("expected packet %d, got %d", consts.JuustagramPacketRangeResp, packetID)
	}
	data := client.Buffer.Bytes()
	if len(data) < 7 {
		t.Fatalf("expected response payload")
	}
	data = data[7:]
	var response protobuf.SC_11706
	if err := proto.Unmarshal(data, &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(response.GetInsMessageList()) != 0 {
		t.Fatalf("expected no messages, got %d", len(response.GetInsMessageList()))
	}
}

func TestJuustagramOpMissingLanguageUsesEmptyText(t *testing.T) {
	initJuustagramHandlerTestDB(t)
	if err := orm.GormDB.Exec("DELETE FROM juustagram_message_states").Error; err != nil {
		t.Fatalf("clear juustagram message state: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM juustagram_templates").Error; err != nil {
		t.Fatalf("clear juustagram templates: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM juustagram_npc_templates").Error; err != nil {
		t.Fatalf("clear juustagram npc templates: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM juustagram_languages").Error; err != nil {
		t.Fatalf("clear juustagram language: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM commanders").Error; err != nil {
		t.Fatalf("clear commanders: %v", err)
	}

	commander := orm.Commander{
		CommanderID: 2001,
		AccountID:   2,
		Level:       1,
		Exp:         0,
		Name:        "Tester",
		LastLogin:   time.Now().UTC(),
	}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}

	message := orm.JuustagramTemplate{
		ID:             10,
		GroupID:        10,
		ShipGroup:      100100,
		Name:           "Tester",
		Sculpture:      "test",
		PicturePersist: "https://example.com/test.png",
		MessagePersist: "ins_missing",
		IsActive:       0,
		TimePersist:    orm.JuustagramTimeConfig{{2024, 1, 1}, {0, 0, 0}},
	}
	if err := orm.GormDB.Create(&message).Error; err != nil {
		t.Fatalf("create juustagram template: %v", err)
	}

	client := &connection.Client{Commander: &commander}
	payload := protobuf.CS_11701{
		Id:  proto.Uint32(10),
		Cmd: proto.Uint32(consts.JuustagramOpActive),
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	_, packetID, err := JuustagramOp(&buffer, client)
	if err != nil {
		t.Fatalf("juustagram op failed: %v", err)
	}
	if packetID != consts.JuustagramPacketOpResponse {
		t.Fatalf("expected packet %d, got %d", consts.JuustagramPacketOpResponse, packetID)
	}
	data := client.Buffer.Bytes()
	if len(data) < 7 {
		t.Fatalf("expected response payload")
	}
	data = data[7:]
	var response protobuf.SC_11702
	if err := proto.Unmarshal(data, &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.GetData() == nil {
		t.Fatalf("expected message data")
	}
	if response.GetData().GetText() != "" {
		t.Fatalf("expected empty message text")
	}
}

func TestJuustagramOpEmptyKeyUsesEmptyText(t *testing.T) {
	initJuustagramHandlerTestDB(t)
	if err := orm.GormDB.Exec("DELETE FROM juustagram_message_states").Error; err != nil {
		t.Fatalf("clear juustagram message state: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM juustagram_templates").Error; err != nil {
		t.Fatalf("clear juustagram templates: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM juustagram_npc_templates").Error; err != nil {
		t.Fatalf("clear juustagram npc templates: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM juustagram_languages").Error; err != nil {
		t.Fatalf("clear juustagram language: %v", err)
	}
	if err := orm.GormDB.Exec("DELETE FROM commanders").Error; err != nil {
		t.Fatalf("clear commanders: %v", err)
	}

	commander := orm.Commander{
		CommanderID: 2002,
		AccountID:   3,
		Level:       1,
		Exp:         0,
		Name:        "Tester",
		LastLogin:   time.Now().UTC(),
	}
	if err := orm.GormDB.Create(&commander).Error; err != nil {
		t.Fatalf("create commander: %v", err)
	}

	message := orm.JuustagramTemplate{
		ID:             11,
		GroupID:        11,
		ShipGroup:      100100,
		Name:           "Tester",
		Sculpture:      "test",
		PicturePersist: "https://example.com/test.png",
		MessagePersist: "",
		IsActive:       0,
		TimePersist:    orm.JuustagramTimeConfig{{2024, 1, 1}, {0, 0, 0}},
	}
	if err := orm.GormDB.Create(&message).Error; err != nil {
		t.Fatalf("create juustagram template: %v", err)
	}

	client := &connection.Client{Commander: &commander}
	payload := protobuf.CS_11701{
		Id:  proto.Uint32(11),
		Cmd: proto.Uint32(consts.JuustagramOpActive),
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	_, packetID, err := JuustagramOp(&buffer, client)
	if err != nil {
		t.Fatalf("juustagram op failed: %v", err)
	}
	if packetID != consts.JuustagramPacketOpResponse {
		t.Fatalf("expected packet %d, got %d", consts.JuustagramPacketOpResponse, packetID)
	}
	data := client.Buffer.Bytes()
	if len(data) < 7 {
		t.Fatalf("expected response payload")
	}
	data = data[7:]
	var response protobuf.SC_11702
	if err := proto.Unmarshal(data, &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.GetData() == nil {
		t.Fatalf("expected message data")
	}
	if response.GetData().GetText() != "" {
		t.Fatalf("expected empty message text")
	}
}

func TestJuustagramComment(t *testing.T) {
	initJuustagramHandlerTestDB(t)
	seedJuustagramHandlerData(t)
	commander := &orm.Commander{CommanderID: 1001}
	client := &connection.Client{Commander: commander}
	payload := protobuf.CS_11703{
		Id:      proto.Uint32(1),
		Discuss: proto.Uint32(1),
		Index:   proto.Uint32(1),
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	_, packetID, err := JuustagramComment(&buffer, client)
	if err != nil {
		t.Fatalf("juustagram comment failed: %v", err)
	}
	if packetID != consts.JuustagramPacketCommentResp {
		t.Fatalf("expected packet %d, got %d", consts.JuustagramPacketCommentResp, packetID)
	}
	data := client.Buffer.Bytes()
	if len(data) < 7 {
		t.Fatalf("expected response payload")
	}
	data = data[7:]
	var response protobuf.SC_11704
	if err := proto.Unmarshal(data, &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.GetData() == nil || len(response.GetData().GetPlayerDiscuss()) != 1 {
		t.Fatalf("expected player discuss entry")
	}
	if response.GetData().GetPlayerDiscuss()[0].GetText() != "player option" {
		t.Fatalf("unexpected player discuss text")
	}
}
