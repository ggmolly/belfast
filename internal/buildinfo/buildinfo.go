package buildinfo

import "strings"

var Commit = "unknown"

func ShortCommit() string {
	trimmed := strings.TrimSpace(Commit)
	if trimmed == "" || strings.EqualFold(trimmed, "unknown") {
		return ""
	}
	if len(trimmed) > 7 {
		return trimmed[:7]
	}
	return trimmed
}
