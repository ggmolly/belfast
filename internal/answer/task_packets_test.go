package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func newTaskPacketTestClient() *connection.Client {
	return &connection.Client{}
}

func TestTaskProgressEventSuccess(t *testing.T) {
	client := newTaskPacketTestClient()
	payload := protobuf.CS_20016{EventType: proto.Uint32(2003), EventTarget: proto.Uint32(1), EventCount: proto.Uint32(2)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	if _, _, err := TaskProgressEvent(&buffer, client); err != nil {
		t.Fatalf("task progress event failed: %v", err)
	}

	var response protobuf.SC_20017
	decodePacketAt(t, client, 0, 20017, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
}

func TestTaskProgressEventRejectsZeroCount(t *testing.T) {
	client := newTaskPacketTestClient()
	payload := protobuf.CS_20016{EventType: proto.Uint32(2003), EventTarget: proto.Uint32(1), EventCount: proto.Uint32(0)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	if _, _, err := TaskProgressEvent(&buffer, client); err != nil {
		t.Fatalf("task progress event failed: %v", err)
	}

	var response protobuf.SC_20017
	decodePacketAt(t, client, 0, 20017, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}

func TestTaskProgressBatchUpdateSuccess(t *testing.T) {
	client := newTaskPacketTestClient()
	payload := protobuf.CS_20009{Progressinfo: []*protobuf.TASK_UPDATE{{Id: proto.Uint32(1001), Mode: proto.Uint32(1), Progress: proto.Uint32(1)}}}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	if _, _, err := TaskProgressBatchUpdate(&buffer, client); err != nil {
		t.Fatalf("task progress batch update failed: %v", err)
	}

	var response protobuf.SC_20010
	decodePacketAt(t, client, 0, 20010, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
}

func TestTaskProgressBatchUpdateRejectsInvalidEntries(t *testing.T) {
	client := newTaskPacketTestClient()
	payload := protobuf.CS_20009{Progressinfo: []*protobuf.TASK_UPDATE{{Id: proto.Uint32(0), Mode: proto.Uint32(1), Progress: proto.Uint32(1)}}}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	if _, _, err := TaskProgressBatchUpdate(&buffer, client); err != nil {
		t.Fatalf("task progress batch update failed: %v", err)
	}

	var response protobuf.SC_20010
	decodePacketAt(t, client, 0, 20010, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}

func TestTaskSubmitSuccess(t *testing.T) {
	client := newTaskPacketTestClient()
	payload := protobuf.CS_20005{Id: proto.Uint32(20820)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	if _, _, err := TaskSubmit(&buffer, client); err != nil {
		t.Fatalf("task submit failed: %v", err)
	}

	var response protobuf.SC_20006
	decodePacketAt(t, client, 0, 20006, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if len(response.GetAwardList()) != 0 {
		t.Fatalf("expected empty awards")
	}
}

func TestTaskSubmitRejectsMissingID(t *testing.T) {
	client := newTaskPacketTestClient()
	payload := protobuf.CS_20005{Id: proto.Uint32(0)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	if _, _, err := TaskSubmit(&buffer, client); err != nil {
		t.Fatalf("task submit failed: %v", err)
	}

	var response protobuf.SC_20006
	decodePacketAt(t, client, 0, 20006, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}

func TestTaskSubmitBatchFiltersInvalidIDs(t *testing.T) {
	client := newTaskPacketTestClient()
	payload := protobuf.CS_20011{IdList: []uint32{20820, 0, 20821}}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	if _, _, err := TaskSubmitBatch(&buffer, client); err != nil {
		t.Fatalf("task submit batch failed: %v", err)
	}

	var response protobuf.SC_20012
	decodePacketAt(t, client, 0, 20012, &response)
	if len(response.GetIdList()) != 2 {
		t.Fatalf("expected 2 task ids, got %d", len(response.GetIdList()))
	}
	if response.GetIdList()[0] != 20820 || response.GetIdList()[1] != 20821 {
		t.Fatalf("unexpected ids: %v", response.GetIdList())
	}
	if len(response.GetAwardList()) != 0 {
		t.Fatalf("expected empty awards")
	}
}

func TestTaskQuickFinishSuccess(t *testing.T) {
	client := newTaskPacketTestClient()
	payload := protobuf.CS_20013{Id: proto.Uint32(20820), ItemCost: proto.Uint32(1)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	if _, _, err := TaskQuickFinish(&buffer, client); err != nil {
		t.Fatalf("task quick finish failed: %v", err)
	}

	var response protobuf.SC_20014
	decodePacketAt(t, client, 0, 20014, &response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
	if len(response.GetAwardList()) != 0 {
		t.Fatalf("expected empty awards")
	}
}

func TestTaskQuickFinishRejectsMissingID(t *testing.T) {
	client := newTaskPacketTestClient()
	payload := protobuf.CS_20013{Id: proto.Uint32(0), ItemCost: proto.Uint32(1)}
	buffer, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	if _, _, err := TaskQuickFinish(&buffer, client); err != nil {
		t.Fatalf("task quick finish failed: %v", err)
	}

	var response protobuf.SC_20014
	decodePacketAt(t, client, 0, 20014, &response)
	if response.GetResult() == 0 {
		t.Fatalf("expected non-zero result")
	}
}
