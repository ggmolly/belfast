package orm

import (
	"sort"
	"time"
)

func OrderedBuilds(builds []Build) []Build {
	if len(builds) == 0 {
		return nil
	}
	ordered := make([]Build, len(builds))
	copy(ordered, builds)
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].ID < ordered[j].ID
	})
	return ordered
}

func RemainingSeconds(finishTime time.Time, now time.Time) uint32 {
	remaining := finishTime.Sub(now)
	if remaining <= 0 {
		return 0
	}
	return uint32(remaining.Seconds())
}
