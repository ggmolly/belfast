package utils

import (
	"fmt"
	"time"
)

// Takes a time.Time object and returns a string representing the time elapsed since then in a human-readable format
func TimeSpan(t time.Time) string {
	diff := time.Since(t)
	if diff.Hours() > 24 {
		return fmt.Sprintf("%d days ago", int(diff.Hours()/24))
	}
	if diff.Hours() > 1 {
		return fmt.Sprintf("%d hours ago", int(diff.Hours()))
	}
	if diff.Minutes() > 1 {
		return fmt.Sprintf("%d minutes ago", int(diff.Minutes()))
	}
	return fmt.Sprintf("%d seconds ago", int(diff.Seconds()))
}
