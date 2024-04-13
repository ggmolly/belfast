package utils

import "time"

// TimeLeft returns a string representing the time left until the given time (hh:mm:ss) or the given string if the time has passed
func TimeLeft(t time.Time, passedString string) string {
	if t.Before(time.Now()) {
		return passedString
	}
	// return in hh:mm:ss format
	return time.Until(t).Round(time.Second).String()
}
