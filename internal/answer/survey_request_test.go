package answer

import (
	"encoding/json"
	"testing"
)

func TestIsSurveyActivityOpenStopMarker(t *testing.T) {
	template := activityTemplate{
		Type:       activityTypeSurvey,
		ConfigID:   42,
		Time:       json.RawMessage(`"stop"`),
		ConfigData: json.RawMessage(`[1, 1]`),
	}
	open, err := isSurveyActivityOpen(template, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if open {
		t.Fatalf("expected survey to be closed for stop marker")
	}
}
