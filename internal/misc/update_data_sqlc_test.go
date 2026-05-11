package misc

import "testing"

func TestApplyBuildTimesSkipsMissingShipTemplate(t *testing.T) {
	buildTimes := map[string]uint32{"10400031": 16200}

	called := false
	skipped, err := applyBuildTimes(buildTimes, func(templateID int64, buildTime int64) (int64, error) {
		called = true
		if templateID != 10400031 {
			t.Fatalf("expected template id 10400031, got %d", templateID)
		}
		if buildTime != 16200 {
			t.Fatalf("expected build time 16200, got %d", buildTime)
		}
		return 0, nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatalf("expected setBuildTime callback to be called")
	}
	if skipped != 1 {
		t.Fatalf("expected one skipped template, got %d", skipped)
	}
}
