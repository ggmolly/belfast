package misc

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ggmolly/belfast/internal/region"
)

func TestGetSpecifiedRegion(t *testing.T) {
	region.ResetCurrentForTest()
	t.Setenv("AL_REGION", "JP")
	if GetSpecifiedRegion() != "JP" {
		t.Fatalf("expected JP region")
	}
}

func TestGetGitHash(t *testing.T) {
	if GetGitHash() == "" {
		t.Fatalf("expected git hash")
	}
}

func TestGetCommits(t *testing.T) {
	commits := GetCommits()
	if len(commits) == 0 {
		t.Fatalf("expected commits to be populated")
	}
}

func TestGetPacketFields(t *testing.T) {
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	defer func() {
		_ = os.Chdir(workingDir)
	}()

	tempDir := t.TempDir()
	protoDir := filepath.Join(tempDir, "packets", "protobuf_src")
	if err := os.MkdirAll(protoDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	protoPath := filepath.Join(protoDir, "test_123.proto")
	content := "required int32 foo = 1;\noptional string bar = 2;"
	if err := os.WriteFile(protoPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write proto: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	fields := GetPacketFields(123)
	if len(fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(fields))
	}
	if fields[0].Name != "foo" || fields[1].Name != "bar" {
		t.Fatalf("unexpected fields: %+v", fields)
	}
}

func TestGetPacketFieldsMissing(t *testing.T) {
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	defer func() {
		_ = os.Chdir(workingDir)
	}()

	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	fields := GetPacketFields(999)
	if len(fields) != 1 || fields[0].Name != "No fields found" {
		t.Fatalf("expected default error fields, got %+v", fields)
	}
}
