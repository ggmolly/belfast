package misc

import (
	"os/exec"
	"strings"
	"time"
)

type Commit struct {
	ShortHash   string
	Author      string
	AuthorEmail string
	Date        time.Time
	Message     string
}

var (
	// Stores last 20 commits
	commits []Commit
)

// GetCommits returns the last 20 commits
func GetCommits() []Commit {
	return commits
}

func init() {
	// Get the last 20 commits
	cmd := exec.Command("git", "log", "--pretty=format:%h|%an|%ae|%ad|%s", "-n", "20")
	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	// Parse the output
	for _, line := range strings.Split(string(out), "\n") {
		commit := Commit{}
		commit.ShortHash = strings.Split(line, "|")[0]
		commit.Author = strings.Split(line, "|")[1]
		commit.AuthorEmail = strings.Split(line, "|")[2]
		commit.Date, _ = time.Parse("Mon Jan 2 15:04:05 2006 -0700", strings.Split(line, "|")[3])
		commit.Message = strings.Split(line, "|")[4]
		commits = append(commits, commit)
	}
}
