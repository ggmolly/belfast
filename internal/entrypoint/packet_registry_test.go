package entrypoint

import (
	"reflect"
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/packets"
)

func TestRegisterTaskPacketClusterHandlers(t *testing.T) {
	original := packets.PacketDecisionFn
	packets.PacketDecisionFn = map[int][]packets.PacketHandler{}
	defer func() {
		packets.PacketDecisionFn = original
	}()

	registerPackets()

	expectSingleHandler(t, 20005, answer.TaskSubmit)
	expectSingleHandler(t, 20009, answer.TaskProgressBatchUpdate)
	expectSingleHandler(t, 20011, answer.TaskSubmitBatch)
	expectSingleHandler(t, 20013, answer.TaskQuickFinish)
	expectSingleHandler(t, 20016, answer.TaskProgressEvent)
}

func expectSingleHandler(t *testing.T, packetID int, expected packets.PacketHandler) {
	t.Helper()
	handlers, ok := packets.PacketDecisionFn[packetID]
	if !ok {
		t.Fatalf("expected packet %d to be registered", packetID)
	}
	if len(handlers) != 1 {
		t.Fatalf("expected packet %d to have 1 handler, got %d", packetID, len(handlers))
	}
	if reflect.ValueOf(handlers[0]).Pointer() != reflect.ValueOf(expected).Pointer() {
		t.Fatalf("expected packet %d to register expected handler", packetID)
	}
}

func TestRegisterPacketsIncludes14004(t *testing.T) {
	packets.PacketDecisionFn = make(map[int][]packets.PacketHandler)
	registerPackets()
	if _, ok := packets.PacketDecisionFn[14004]; !ok {
		t.Fatalf("expected handler for CS_14004 to be registered")
	}
}

func TestRegisterPacketsIncludesLoveLetterGetAll(t *testing.T) {
	packets.PacketDecisionFn = make(map[int][]packets.PacketHandler)
	registerPackets()
	if _, ok := packets.PacketDecisionFn[12406]; !ok {
		t.Fatalf("expected handler for CS_12406 to be registered")
	}
}
