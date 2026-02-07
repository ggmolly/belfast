package answer

import (
	"testing"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestExercisePowerRankList_Success_ReturnsSC18007(t *testing.T) {
	client := setupExerciseTest(t)

	payload := protobuf.CS_18006{Type: proto.Uint32(0)}
	buf, err := proto.Marshal(&payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if _, _, err := ExercisePowerRankList(&buf, client); err != nil {
		t.Fatalf("ExercisePowerRankList failed: %v", err)
	}

	var resp protobuf.SC_18007
	decodePacketMessage(t, client, 18007, &resp)
	if len(resp.GetArenaRankLsit()) != 5 {
		t.Fatalf("expected 5 rank entries, got %d", len(resp.GetArenaRankLsit()))
	}
	first := resp.GetArenaRankLsit()[0]
	if first.GetId() != client.Commander.CommanderID {
		t.Fatalf("expected first id to be commander id")
	}
	if first.GetName() != client.Commander.Name {
		t.Fatalf("expected first name to be commander name")
	}
}

func TestExercisePowerRankList_InvalidPayload_ReturnsError(t *testing.T) {
	client := setupExerciseTest(t)
	buf := []byte{0xff}
	if _, _, err := ExercisePowerRankList(&buf, client); err == nil {
		t.Fatalf("expected error")
	}
	if client.Buffer.Len() != 0 {
		t.Fatalf("expected no response to be written")
	}
}
