package answer

import (
	"encoding/json"
	"testing"
	"time"
)

func TestBoolToUint32(t *testing.T) {
	if boolToUint32(true) != 1 {
		t.Fatalf("expected true to map to 1")
	}
	if boolToUint32(false) != 0 {
		t.Fatalf("expected false to map to 0")
	}
}

func TestTBPlaceholders(t *testing.T) {
	info := tbInfoPlaceholder()
	if info == nil || info.GetFsm() == nil || info.GetRound() == nil {
		t.Fatalf("expected TBINFO placeholder to be populated")
	}
	permanent := tbPermanentPlaceholder()
	if permanent == nil || permanent.GetNgPlusCount() == 0 {
		t.Fatalf("expected TBPERMANENT placeholder to be populated")
	}
}

func TestServerTicketRoundTrip(t *testing.T) {
	ticket := formatServerTicket(12345)
	if parseServerTicket(ticket) != 12345 {
		t.Fatalf("expected to parse arg2 from ticket")
	}
	if parseServerTicket(formatServerTicket(0)) != 0 {
		t.Fatalf("expected zero arg2 for prefix-only ticket")
	}
}

func TestParseServerTicketInvalid(t *testing.T) {
	if parseServerTicket("invalid") != 0 {
		t.Fatalf("expected invalid ticket to return 0")
	}
	if parseServerTicket(serverTicketPrefix+":not-a-number") != 0 {
		t.Fatalf("expected non-numeric ticket to return 0")
	}
}

func TestActivityStopTimeValid(t *testing.T) {
	raw := json.RawMessage(`["timer",0,[[2026,1,2],[3,4,5]]]`)
	result := activityStopTime(raw)
	stop := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	if result != uint32(stop.Unix()) {
		t.Fatalf("expected stop time to match")
	}
}

func TestActivityStopTimeInvalid(t *testing.T) {
	cases := []json.RawMessage{
		json.RawMessage(`"label"`),
		json.RawMessage(`["timer"]`),
		json.RawMessage(`["not-timer",0,[]]`),
		json.RawMessage(`["timer",0,123]`),
		json.RawMessage(`["timer",0,[[2026],[3,4,5]]]`),
		json.RawMessage(`["timer",0,[[2026,1,2],[3,4]]]`),
		json.RawMessage(`["timer",0,[["bad",1,2],[3,4,5]]]`),
	}
	for _, raw := range cases {
		if activityStopTime(raw) != 0 {
			t.Fatalf("expected invalid stop time to return 0")
		}
	}
}

func TestParseJSONInt(t *testing.T) {
	if value, ok := parseJSONInt(3.0); !ok || value != 3 {
		t.Fatalf("expected float64 value to parse")
	}
	if _, ok := parseJSONInt("bad"); ok {
		t.Fatalf("expected non-number to fail")
	}
}

func TestParseJSONUint(t *testing.T) {
	if value, ok := parseJSONUint(3.0); !ok || value != 3 {
		t.Fatalf("expected float64 value to parse")
	}
	if value, ok := parseJSONUint(5); !ok || value != 5 {
		t.Fatalf("expected int value to parse")
	}
	if value, ok := parseJSONUint(uint32(7)); !ok || value != 7 {
		t.Fatalf("expected uint32 value to parse")
	}
	if _, ok := parseJSONUint("bad"); ok {
		t.Fatalf("expected non-number to fail")
	}
}

func TestParseActivityConfigIDs(t *testing.T) {
	data, err := parseActivityConfigIDs(json.RawMessage(`[1,2,3]`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) != 3 || data[0] != 1 || data[2] != 3 {
		t.Fatalf("unexpected ids: %v", data)
	}

	data, err = parseActivityConfigIDs(json.RawMessage(`[1,"2",3]`))
	if err == nil {
		t.Fatalf("expected error for unsupported id")
	}
}
