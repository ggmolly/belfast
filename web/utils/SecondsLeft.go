package utils

import "time"

// Returns the number of seconds left until the given time
// returns 0 if the given time is in the past
func SecondsLeft(t time.Time) int {
	seconds := int(time.Until(t).Seconds())
	if seconds < 0 {
		return 0
	}
	return seconds
}
