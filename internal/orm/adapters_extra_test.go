package orm

import (
	"testing"
	"time"
)

func TestToProtoBuildInfoUsesPayloadFields(t *testing.T) {
	finish := time.Unix(123, 0)
	payload := BuildInfoPayload{
		Build:      nil,
		PoolID:     7,
		BuildTime:  45,
		FinishTime: finish,
	}
	info := ToProtoBuildInfo(payload)
	if info.GetTime() != 45 {
		t.Fatalf("expected build time 45, got %d", info.GetTime())
	}
	if info.GetFinishTime() != 123 {
		t.Fatalf("expected finish time 123, got %d", info.GetFinishTime())
	}
	if info.GetBuildId() != 7 {
		t.Fatalf("expected pool id 7, got %d", info.GetBuildId())
	}
}

func TestToProtoBuildInfoUsesBuildOverrides(t *testing.T) {
	finish := time.Unix(500, 0)
	build := Build{PoolID: 9, FinishesAt: finish}
	payload := BuildInfoPayload{
		Build:      &build,
		PoolID:     7,
		BuildTime:  12,
		FinishTime: time.Unix(100, 0),
	}
	info := ToProtoBuildInfo(payload)
	if info.GetFinishTime() != 500 {
		t.Fatalf("expected finish time 500, got %d", info.GetFinishTime())
	}
	if info.GetBuildId() != 9 {
		t.Fatalf("expected pool id 9, got %d", info.GetBuildId())
	}
}

func TestBoolToUint32(t *testing.T) {
	if boolToUint32(true) != 1 {
		t.Fatalf("expected true to map to 1")
	}
	if boolToUint32(false) != 0 {
		t.Fatalf("expected false to map to 0")
	}
}

func TestToProtoDropInfoList(t *testing.T) {
	attachments := []MailAttachment{{Type: 1, ItemID: 2, Quantity: 3}}
	drops := ToProtoDropInfoList(attachments)
	if len(drops) != 1 {
		t.Fatalf("expected 1 drop, got %d", len(drops))
	}
	if drops[0].GetType() != 1 || drops[0].GetId() != 2 || drops[0].GetNumber() != 3 {
		t.Fatalf("unexpected drop values: %+v", drops[0])
	}
}

func TestToProtoCompensationDropInfoList(t *testing.T) {
	attachments := []CompensationAttachment{{Type: 2, ItemID: 4, Quantity: 6}}
	drops := ToProtoCompensationDropInfoList(attachments)
	if len(drops) != 1 {
		t.Fatalf("expected 1 drop, got %d", len(drops))
	}
	if drops[0].GetType() != 2 || drops[0].GetId() != 4 || drops[0].GetNumber() != 6 {
		t.Fatalf("unexpected drop values: %+v", drops[0])
	}
}

func TestToProtoProposeResponse(t *testing.T) {
	success := ToProtoProposeResponse(true)
	if success.GetResult() != 0 {
		t.Fatalf("expected success result 0, got %d", success.GetResult())
	}
	failure := ToProtoProposeResponse(false)
	if failure.GetResult() != 1 {
		t.Fatalf("expected failure result 1, got %d", failure.GetResult())
	}
	if success.GetTime() == 0 || failure.GetTime() == 0 {
		t.Fatalf("expected timestamps to be set")
	}
}
