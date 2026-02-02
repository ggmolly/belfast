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
	payload := &protobuf.CS_11513{ShopId: proto.Uint32(1), Device: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.RefundChargeCommandAnswer(&buf, client); err != nil {
		t.Fatalf("RefundChargeCommandAnswer failed: %v", err)
	}
	response := &protobuf.SC_11514{}
	decodeTestPacket(t, client, 11514, response)
	if response.GetResult() != 5002 {
		t.Fatalf("expected result 5002, got %d", response.GetResult())
	}
	if response.GetPayId() != "" || response.GetUrl() != "" || response.GetOrderSign() != "" {
		t.Fatalf("expected empty payment fields")
	}
}

func TestChargeConfirmDisabled(t *testing.T) {
	client := &connection.Client{}
	payload := &protobuf.CS_11504{PayId: proto.String("pay-1"), PayIdBili: proto.String("bili-1")}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.ChargeConfirmCommandAnswer(&buf, client); err != nil {
		t.Fatalf("ChargeConfirmCommandAnswer failed: %v", err)
	}
	response := &protobuf.SC_11505{}
	decodeTestPacket(t, client, 11505, response)
	if response.GetResult() != 5002 {
		t.Fatalf("expected result 5002, got %d", response.GetResult())
	}
	if response.GetShopId() != 0 || response.GetGem() != 0 || response.GetGemFree() != 0 {
		t.Fatalf("expected zeroed shop and gem fields")
	}
}

func TestChargeFailedAck(t *testing.T) {
	client := &connection.Client{}
	payload := &protobuf.CS_11510{PayId: proto.String("pay-1"), Code: proto.Uint32(1)}
	buf, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if _, _, err := answer.ChargeFailedCommandAnswer(&buf, client); err != nil {
		t.Fatalf("ChargeFailedCommandAnswer failed: %v", err)
	}
	response := &protobuf.SC_11511{}
	decodeTestPacket(t, client, 11511, response)
	if response.GetResult() != 0 {
		t.Fatalf("expected result 0, got %d", response.GetResult())
	}
}
