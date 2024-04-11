package utils

import (
	"fmt"
	"time"
)

// Takes a time.Time object and returns a string representing the time elapsed since then in a human-readable format
// dd/mm/yyyy hh:mm:ss
func TimeFormat(t time.Time) string {
	return fmt.Sprintf("%02d/%02d/%d %02d:%02d:%02d", t.Day(), t.Month(), t.Year(), t.Hour(), t.Minute(), t.Second())
}
