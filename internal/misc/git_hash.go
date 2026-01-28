package misc

import "github.com/ggmolly/belfast/internal/buildinfo"

// GetGitHash returns the git hash of the current build
func GetGitHash() string {
	return buildinfo.ShortCommit()
}
