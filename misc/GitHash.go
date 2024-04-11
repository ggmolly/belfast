package misc

import (
	"os/exec"
	"strings"
)

var (
	gitHash string = "fffffff"
)

// GetGitHash returns the git hash of the current build
func GetGitHash() string {
	return gitHash
}

func init() {
	shortHash, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return
	}
	gitHash = strings.TrimSpace(string(shortHash))
}
