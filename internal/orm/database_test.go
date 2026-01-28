package orm

import (
	"os"
	"testing"
)

func TestInitDatabase(t *testing.T) {
	originalDB := GormDB
	defer func() {
		GormDB = originalDB
	}()

	os.Setenv("MODE", "test")
	defer os.Setenv("MODE", "")

	if !InitDatabase() {
		t.Fatalf("expected InitDatabase to return true")
	}

	if GormDB == nil {
		t.Fatalf("expected GormDB to be initialized")
	}
}

func TestInitSqlite(t *testing.T) {
	db := initSqlite("file::memory:?cache=shared")

	if db == nil {
		t.Fatalf("expected initSqlite to return non-nil db")
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get underlying sql.DB: %v", err)
	}

	if sqlDB == nil {
		t.Fatalf("expected underlying sql.DB to be non-nil")
	}

	sqlDB.Close()
}

func TestInitDatabasePanicOnInvalidDSN(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected InitDatabase to panic on invalid DSN")
		}
	}()

	initSqlite("file:///invalid/path/that/does/not/exist.db")
}

func TestInitDatabaseTestMode(t *testing.T) {
	originalDB := GormDB
	defer func() {
		GormDB = originalDB
	}()

	os.Setenv("MODE", "test")
	defer os.Setenv("MODE", "")

	result := InitDatabase()

	if !result {
		t.Fatalf("expected InitDatabase to return true in test mode")
	}
}

func TestInitDatabaseProductionMode(t *testing.T) {
	originalDB := GormDB
	originalMode := os.Getenv("MODE")
	defer func() {
		GormDB = originalDB
		os.Setenv("MODE", originalMode)
	}()

	os.Setenv("MODE", "")
	defer os.Setenv("MODE", "test")

	result := InitDatabase()

	if GormDB == nil {
		t.Fatalf("expected GormDB to be initialized")
	}

	if result {
		t.Fatalf("expected InitDatabase to return false when database is already seeded")
	}
}

func TestAutoMigrate(t *testing.T) {
	originalDB := GormDB
	defer func() {
		GormDB = originalDB
	}()

	os.Setenv("MODE", "test")
	defer os.Setenv("MODE", "")

	InitDatabase()

	if err := GormDB.AutoMigrate(&Server{}); err != nil {
		t.Fatalf("expected AutoMigrate to succeed, got error: %v", err)
	}
}

func TestSeedDatabaseSkip(t *testing.T) {
	originalDB := GormDB
	defer func() {
		GormDB = originalDB
	}()

	os.Setenv("MODE", "test")
	defer os.Setenv("MODE", "")

	if !seedDatabase(true) {
		t.Fatalf("expected seedDatabase to return true when skipSeed is true")
	}
}

func TestGormDBGlobalVariable(t *testing.T) {
	t.Setenv("MODE", "test")
	defer os.Setenv("MODE", "")

	InitDatabase()

	originalDB := GormDB
	defer func() {
		GormDB = originalDB
	}()

	GormDB = nil

	if GormDB != nil {
		t.Fatalf("expected GormDB to be nil after assignment to nil")
	}

	GormDB = originalDB

	if GormDB == nil {
		t.Fatalf("expected GormDB to be non-nil after restoring original")
	}
}

func TestInitDatabaseCreatesGormDBInstance(t *testing.T) {
	originalDB := GormDB
	defer func() {
		GormDB = originalDB
	}()

	os.Setenv("MODE", "test")
	defer os.Setenv("MODE", "")

	InitDatabase()

	if GormDB == nil {
		t.Fatalf("expected GormDB to be non-nil after InitDatabase")
	}
}

func TestInitDatabasePreparesStatements(t *testing.T) {
	db := initSqlite("file::memory:?cache=shared")

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get underlying sql.DB: %v", err)
	}

	if sqlDB == nil {
		t.Fatalf("expected underlying sql.DB to be non-nil")
	}

	sqlDB.Close()

	if db.Statement == nil {
		t.Fatalf("expected db.Statement to be initialized")
	}
}
