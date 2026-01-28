package response

import "testing"

func TestSuccess(t *testing.T) {
	data := map[string]string{"key": "value"}
	payload := Success(data)
	if !payload.OK {
		t.Fatalf("expected ok true")
	}
	if payload.Data == nil {
		t.Fatalf("expected data to be set")
	}
	if payload.Error != nil {
		t.Fatalf("expected error to be nil")
	}
}

func TestError(t *testing.T) {
	payload := Error("bad_request", "invalid", map[string]string{"field": "name"})
	if payload.OK {
		t.Fatalf("expected ok false")
	}
	if payload.Error == nil {
		t.Fatalf("expected error to be set")
	}
	if payload.Error.Code != "bad_request" || payload.Error.Message != "invalid" {
		t.Fatalf("unexpected error payload: %+v", payload.Error)
	}
}
