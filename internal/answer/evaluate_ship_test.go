package answer

import (
	"strings"
	"sync"
	"testing"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

var shipEvaluationTestOnce sync.Once

func initShipEvaluationTestDB(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	shipEvaluationTestOnce.Do(func() {
		orm.InitDatabase()
	})
}

func resetShipEvaluationState(t *testing.T) {
	t.Helper()
	shipDiscussStoreMu.Lock()
	shipDiscussStore = map[uint32]*shipDiscussState{}
	shipDiscussStoreMu.Unlock()

	if err := orm.GormDB.Exec("DELETE FROM likes").Error; err != nil {
		t.Fatalf("clear likes: %v", err)
	}
}

func decodeShipEvaluationResponse(t *testing.T, client *connection.Client) protobuf.SC_17104 {
	t.Helper()
	data := client.Buffer.Bytes()
	if len(data) < 7 {
		t.Fatalf("expected response payload")
	}
	data = data[7:]
	var response protobuf.SC_17104
	if err := proto.Unmarshal(data, &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	return response
}

func TestPostShipEvaluationCommentSuccessReturnsShipDiscuss(t *testing.T) {
	initShipEvaluationTestDB(t)
	resetShipEvaluationState(t)

	commander := &orm.Commander{CommanderID: 1, Name: "Tester", Level: 1}
	client := &connection.Client{Commander: commander}

	payload := protobuf.CS_17103{
		ShipGroupId: proto.Uint32(1010),
		Context:     proto.String("hello world"),
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	_, packetID, err := PostShipEvaluationComment(&buffer, client)
	if err != nil {
		t.Fatalf("post ship evaluation comment failed: %v", err)
	}
	if packetID != 17104 {
		t.Fatalf("expected packet 17104, got %d", packetID)
	}

	response := decodeShipEvaluationResponse(t, client)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if response.NeedLevel == nil {
		t.Fatalf("expected need_level to be present")
	}
	if response.GetShipDiscuss() == nil {
		t.Fatalf("expected ship_discuss to be non-nil on success")
	}
	if response.GetShipDiscuss().GetShipGroupId() != 1010 {
		t.Fatalf("expected ship group 1010, got %d", response.GetShipDiscuss().GetShipGroupId())
	}
	if response.GetShipDiscuss().GetDiscussCount() != 1 {
		t.Fatalf("expected discuss count 1, got %d", response.GetShipDiscuss().GetDiscussCount())
	}
	if response.GetShipDiscuss().GetDailyDiscussCount() != 1 {
		t.Fatalf("expected daily discuss count 1, got %d", response.GetShipDiscuss().GetDailyDiscussCount())
	}
	if len(response.GetShipDiscuss().GetDiscussList()) != 1 {
		t.Fatalf("expected 1 discuss entry, got %d", len(response.GetShipDiscuss().GetDiscussList()))
	}
	entry := response.GetShipDiscuss().GetDiscussList()[0]
	if entry.GetContext() != "hello world" {
		t.Fatalf("expected context to match")
	}
	if entry.GetNickName() != "Tester" {
		t.Fatalf("expected nick_name to match")
	}
	if entry.GetId() == 0 {
		t.Fatalf("expected discuss id to be non-zero")
	}
}

func TestPostShipEvaluationCommentTooLongReturns2011(t *testing.T) {
	initShipEvaluationTestDB(t)
	resetShipEvaluationState(t)

	client := &connection.Client{Commander: &orm.Commander{CommanderID: 1, Name: "Tester", Level: 1}}

	tooLong := strings.Repeat("a", shipEvaluationCommentMaxRunes+1)
	payload := protobuf.CS_17103{
		ShipGroupId: proto.Uint32(2020),
		Context:     proto.String(tooLong),
	}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	_, packetID, err := PostShipEvaluationComment(&buffer, client)
	if err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	if packetID != 17104 {
		t.Fatalf("expected packet 17104, got %d", packetID)
	}
	response := decodeShipEvaluationResponse(t, client)
	if response.GetResult() != 2011 {
		t.Fatalf("expected result 2011, got %d", response.GetResult())
	}
	if response.NeedLevel == nil {
		t.Fatalf("expected need_level to be present")
	}
	if response.GetShipDiscuss() != nil {
		t.Fatalf("expected ship_discuss to be nil on error")
	}
}

func TestPostShipEvaluationCommentEnforcesDailyLimit(t *testing.T) {
	initShipEvaluationTestDB(t)
	resetShipEvaluationState(t)

	client := &connection.Client{Commander: &orm.Commander{CommanderID: 1, Name: "Tester", Level: 1}}
	shipGroupID := uint32(3030)

	for i := 0; i < shipEvaluationDailyCommentMax; i++ {
		payload := protobuf.CS_17103{
			ShipGroupId: proto.Uint32(shipGroupID),
			Context:     proto.String("ok"),
		}
		buffer, err := proto.Marshal(&payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		_, _, err = PostShipEvaluationComment(&buffer, client)
		if err != nil {
			t.Fatalf("handler failed: %v", err)
		}
		client.Buffer.Reset()
	}

	last := protobuf.CS_17103{ShipGroupId: proto.Uint32(shipGroupID), Context: proto.String("excess")}
	buffer, err := proto.Marshal(&last)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	_, _, err = PostShipEvaluationComment(&buffer, client)
	if err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	response := decodeShipEvaluationResponse(t, client)
	if response.GetResult() != 1 {
		t.Fatalf("expected result 1 for daily limit, got %d", response.GetResult())
	}
	if response.NeedLevel == nil {
		t.Fatalf("expected need_level to be present")
	}
}
