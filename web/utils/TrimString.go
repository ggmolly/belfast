package utils

import "strings"

func TrimString(s string, n int) string {
	if len(s) > n {
		return strings.Trim(s[:n-3], " ") + "..."
	}
	return s
}
