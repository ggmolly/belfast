package answer_test

import (
	"testing"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestRefundChargeDisabled(t *testing.T) {
	client := &connection.Client{}
	payload := &protobuf.CS_11513{
		ShopId: proto.Uint32(1),
		Device: proto.Uint32(2),
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.RefundChargeAnswer(&buf, client); err != nil {
		t.Fatalf("RefundChargeAnswer failed: %v", err)
	}
	response := &protobuf.SC_11514{}
	decodeTestPacket(t, client, 11514, response)
	if response.GetResult() != 5002 {
		t.Fatalf("expected result 5002, got %d", response.GetResult())
	}
	if response.GetPayId() != "" {
		t.Fatalf("expected pay_id empty, got %q", response.GetPayId())
	}
	if response.GetUrl() != "" {
		t.Fatalf("expected url empty, got %q", response.GetUrl())
	}
	if response.GetOrderSign() != "" {
		t.Fatalf("expected order_sign empty, got %q", response.GetOrderSign())
	}
}
