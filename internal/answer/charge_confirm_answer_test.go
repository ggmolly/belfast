package answer_test

import (
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestChargeConfirmDisabled(t *testing.T) {
	client := &connection.Client{}
	payload := &protobuf.CS_11504{
		PayId:     proto.String("pay-11504"),
		PayIdBili: proto.String("bili-11504"),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.ChargeConfirmAnswer(&buf, client); err != nil {
		t.Fatalf("ChargeConfirmAnswer failed: %v", err)
	}
	response := &protobuf.SC_11505{}
	decodeTestPacket(t, client, 11505, response)
	if response.GetResult() != 5002 {
		t.Fatalf("expected result 5002, got %d", response.GetResult())
	}
	if response.GetShopId() != 0 {
		t.Fatalf("expected shop_id 0, got %d", response.GetShopId())
	}
	if response.GetGem() != 0 {
		t.Fatalf("expected gem 0, got %d", response.GetGem())
	}
	if response.GetGemFree() != 0 {
		t.Fatalf("expected gem_free 0, got %d", response.GetGemFree())
	}
}
