package orm

import (
	"os"
	"sync"
	"testing"
)

func TestInitDatabasePanicsWithoutDSN(t *testing.T) {
	defer func() {
		initOnce = sync.Once{}
		initErr = nil
	}()

	initOnce = sync.Once{}
	initErr = nil
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(cwd) }()
	workdir := t.TempDir()
	if err := os.Chdir(workdir); err != nil {
		t.Fatalf("failed to switch to temp dir: %v", err)
	}
	t.Setenv("BELFAST_TEST_POSTGRES_DSN", "")
	t.Setenv("TEST_DATABASE_DSN", "")
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected InitDatabase to panic without DSN")
		}
	}()
	InitDatabase()
}
