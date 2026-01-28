package api

import "testing"

func TestStartDisabled(t *testing.T) {
	if err := Start(Config{Enabled: false}); err != nil {
		t.Fatalf("expected nil when disabled, got %v", err)
	}
}
