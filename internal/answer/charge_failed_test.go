package answer_test

import (
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestChargeFailedAck(t *testing.T) {
	client := &connection.Client{}
	payload := &protobuf.CS_11510{
		PayId: proto.String("pay-11510"),
		Code:  proto.Uint32(100),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.ChargeFailed(&buf, client); err != nil {
		t.Fatalf("ChargeFailed failed: %v", err)
	}
	response := &protobuf.SC_11511{}
	decodeTestPacket(t, client, 11511, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
}
