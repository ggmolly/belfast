package orm

import (
	"testing"
	"time"
)

func TestOrderedBuildsEmpty(t *testing.T) {
	if OrderedBuilds(nil) != nil {
		t.Fatalf("expected nil for empty builds")
	}
}

func TestOrderedBuildsSortsCopy(t *testing.T) {
	builds := []Build{{ID: 3}, {ID: 1}, {ID: 2}}
	ordered := OrderedBuilds(builds)
	if ordered[0].ID != 1 || ordered[1].ID != 2 || ordered[2].ID != 3 {
		t.Fatalf("unexpected order: %+v", ordered)
	}
	if builds[0].ID != 3 {
		t.Fatalf("expected original slice unchanged")
	}
}

func TestRemainingSeconds(t *testing.T) {
	now := time.Unix(100, 0)
	if RemainingSeconds(now.Add(-time.Second), now) != 0 {
		t.Fatalf("expected 0 for past finish")
	}
	if RemainingSeconds(now, now) != 0 {
		t.Fatalf("expected 0 for same time")
	}
	if RemainingSeconds(now.Add(5*time.Second), now) != 5 {
		t.Fatalf("expected 5 seconds remaining")
	}
}
