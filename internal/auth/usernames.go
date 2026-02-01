package auth

import "strings"

func NormalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}
