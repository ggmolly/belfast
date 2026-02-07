package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestExerciseReplaceRivals_Success_ReturnsSC18004(t *testing.T) {
	client := setupConfigTest(t)

	payload := protobuf.CS_18003{Type: proto.Uint32(0)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ExerciseReplaceRivals(&buf, client); err != nil {
		t.Fatalf("ExerciseReplaceRivals failed: %v", err)
	}

	var resp protobuf.SC_18004
	decodePacketMessage(t, client, 18004, &resp)
	if resp.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", resp.GetResult())
	}
	if len(resp.GetTargetList()) != 5 {
		t.Fatalf("expected 5 rivals, got %d", len(resp.GetTargetList()))
	}
	for _, rival := range resp.GetTargetList() {
		if rival.GetId() == 0 || rival.GetLevel() == 0 || rival.GetName() == "" {
			t.Fatalf("expected rival required fields to be set")
		}
	}
}

func TestExerciseReplaceRivals_InvalidPayload_ReturnsError(t *testing.T) {
	client := setupConfigTest(t)
	buf := []byte{0xff}
	if _, _, err := ExerciseReplaceRivals(&buf, client); err == nil {
		t.Fatalf("expected error")
	}
	if client.Buffer.Len() != 0 {
		t.Fatalf("expected no response to be written")
	}
}
