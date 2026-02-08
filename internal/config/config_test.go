package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	configContent := `[belfast]
bind_address = "127.0.0.1"
port = 8080
maintenance = false
name = "Test Server"

[api]
enabled = true
port = 8081
environment = "production"
cors_origins = ["http://localhost:8080"]

[database]
path = "belfast.db"
schema_name = "custom_schema"

[region]
default = "EN"

[create_player]
skip_onboarding = true
name_blacklist = ["test", "banned"]
name_illegal_pattern = "^admin$"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Belfast.BindAddress != "127.0.0.1" {
		t.Fatalf("expected bind address '127.0.0.1', got %s", cfg.Belfast.BindAddress)
	}
	if cfg.Belfast.Port != 8080 {
		t.Fatalf("expected port 8080, got %d", cfg.Belfast.Port)
	}
	if cfg.Belfast.Maintenance != false {
		t.Fatalf("expected maintenance false, got %v", cfg.Belfast.Maintenance)
	}
	if cfg.Belfast.Name != "Test Server" {
		t.Fatalf("expected name 'Test Server', got %s", cfg.Belfast.Name)
	}
	if cfg.API.Enabled != true {
		t.Fatalf("expected api enabled true, got %v", cfg.API.Enabled)
	}
	if cfg.API.Environment != "production" {
		t.Fatalf("expected environment 'production', got %s", cfg.API.Environment)
	}
	if len(cfg.API.CORSOrigins) != 1 {
		t.Fatalf("expected 1 cors origin, got %d", len(cfg.API.CORSOrigins))
	}
	if cfg.API.CORSOrigins[0] != "http://localhost:8080" {
		t.Fatalf("expected cors origin 'http://localhost:8080', got %s", cfg.API.CORSOrigins[0])
	}
	if cfg.DB.Path != "custom_schema.db" {
		t.Fatalf("expected database path 'custom_schema.db', got %s", cfg.DB.Path)
	}
	if cfg.Region.Default != "EN" {
		t.Fatalf("expected default region 'EN', got %s", cfg.Region.Default)
	}
	if cfg.CreatePlayer.SkipOnboarding != true {
		t.Fatalf("expected skip_onboarding true, got %v", cfg.CreatePlayer.SkipOnboarding)
	}
	if len(cfg.CreatePlayer.NameBlacklist) != 2 {
		t.Fatalf("expected 2 blacklisted names, got %d", len(cfg.CreatePlayer.NameBlacklist))
	}
	if cfg.CreatePlayer.NameIllegalPattern != "^admin$" {
		t.Fatalf("expected name_illegal_pattern '^admin$', got %s", cfg.CreatePlayer.NameIllegalPattern)
	}
	if cfg.Path != configPath {
		t.Fatalf("expected path to be set to %s, got %s", configPath, cfg.Path)
	}
}

func TestLoadDefaultPort(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	configContent := `[belfast]
bind_address = "127.0.0.1"
maintenance = false

[api]
enabled = false

[database]
path = "test.db"

[region]
default = "CN"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Belfast.Port != 80 {
		t.Fatalf("expected default port 80, got %d", cfg.Belfast.Port)
	}
}

func TestLoadDefaultDatabasePath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	configContent := `[belfast]
bind_address = "127.0.0.1"
port = 8080

[api]
enabled = false

[database]
schema_name = ""

[region]
default = "US"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.DB.Path != "data/belfast.db" {
		t.Fatalf("expected default db path 'data/belfast.db', got %s", cfg.DB.Path)
	}
}

func TestLoadSchemaNameFromBelfastName(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	configContent := `[belfast]
bind_address = "127.0.0.1"
port = 8080
name = "My Server 2"

[api]
enabled = false

[database]
path = "data/belfast.db"

[region]
default = "US"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.DB.Path != filepath.Join("data", "my_server_2.db") {
		t.Fatalf("expected schema-based db path, got %s", cfg.DB.Path)
	}
}

func TestLoadSchemaNameOverridesBelfastName(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	configContent := `[belfast]
bind_address = "127.0.0.1"
port = 8080
name = "Ignored Name"

[api]
enabled = false

[database]
path = "data/belfast.db"
schema_name = "custom_schema"

[region]
default = "US"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.DB.Path != filepath.Join("data", "custom_schema.db") {
		t.Fatalf("expected schema_name to override name, got %s", cfg.DB.Path)
	}
}

func TestLoadPostgresSchemaNameDoesNotRewritePath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	configContent := `[belfast]
bind_address = "127.0.0.1"
port = 8080
name = "My Server"

[api]
enabled = false

[database]
driver = "postgres"
path = "data/belfast.db"
dsn = "postgres://user:pass@localhost:5432/belfast?sslmode=disable"
schema_name = "custom_schema"

[region]
default = "US"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if cfg.DB.Driver != "postgres" {
		t.Fatalf("expected db driver 'postgres', got %s", cfg.DB.Driver)
	}
	if cfg.DB.SchemaName != "custom_schema" {
		t.Fatalf("expected schema_name 'custom_schema', got %s", cfg.DB.SchemaName)
	}
	if cfg.DB.Path != "data/belfast.db" {
		t.Fatalf("expected db path to remain unchanged, got %s", cfg.DB.Path)
	}
}

func TestLoadMySQLSchemaNameDoesNotRewritePath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	configContent := `[belfast]
bind_address = "127.0.0.1"
port = 8080
name = "My Server"

[api]
enabled = false

[database]
driver = "mysql"
path = "data/belfast.db"
dsn = "user:pass@tcp(localhost:3306)/belfast?parseTime=true&charset=utf8mb4"
schema_name = "custom_schema"

[region]
default = "US"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if cfg.DB.Driver != "mysql" {
		t.Fatalf("expected db driver 'mysql', got %s", cfg.DB.Driver)
	}
	if cfg.DB.SchemaName != "custom_schema" {
		t.Fatalf("expected schema_name 'custom_schema', got %s", cfg.DB.SchemaName)
	}
	if cfg.DB.Path != "data/belfast.db" {
		t.Fatalf("expected db path to remain unchanged, got %s", cfg.DB.Path)
	}
}

func TestLoadMissingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "missing.toml")

	_, err := Load(configPath)
	if err == nil {
		t.Fatalf("expected error for missing config file")
	}
	expectedErr := "config file missing"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Fatalf("expected error containing %q, got %q", expectedErr, err.Error())
	}
}

func TestLoadGatewayConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "gateway.toml")
	configContent := `bind_address = "127.0.0.1"
port = 8088

[[servers]]
id = 1
ip = "127.0.0.1"
port = 7000
api_port = 2289
assert_online = true
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfg, err := LoadGateway(configPath)
	if err != nil {
		t.Fatalf("failed to load gateway config: %v", err)
	}
	if cfg.BindAddress != "127.0.0.1" {
		t.Fatalf("expected bind address '127.0.0.1', got %s", cfg.BindAddress)
	}
	if cfg.Port != 8088 {
		t.Fatalf("expected port 8088, got %d", cfg.Port)
	}
	if len(cfg.Servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(cfg.Servers))
	}
	if !cfg.Servers[0].AssertOnline {
		t.Fatalf("expected server assert_online to be true")
	}
	if cfg.Servers[0].ID != 1 {
		t.Fatalf("expected server id 1, got %d", cfg.Servers[0].ID)
	}
	if Current().Belfast.Port != 8088 {
		t.Fatalf("expected current port 8088, got %d", Current().Belfast.Port)
	}
}

func TestLoadGatewayDefaultPort(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "gateway.toml")
	configContent := `bind_address = "127.0.0.1"

[[servers]]
id = 1
ip = "127.0.0.1"
port = 7000
api_port = 2289
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfg, err := LoadGateway(configPath)
	if err != nil {
		t.Fatalf("failed to load gateway config: %v", err)
	}
	if cfg.Port != 80 {
		t.Fatalf("expected default gateway port 80, got %d", cfg.Port)
	}
}

func TestLoadGatewayMissingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "missing.toml")

	_, err := LoadGateway(configPath)
	if err == nil {
		t.Fatalf("expected error for missing config file")
	}
	if !strings.Contains(err.Error(), "config file missing") {
		t.Fatalf("expected missing config error, got %q", err.Error())
	}
}

func TestLoadGatewayInvalidToml(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.toml")
	configContent := `bind_address = "127.0.0.1"
port = 8088

[[servers]]
id = 1
ip = "127.0.0.1"
port = 7000
api_port =
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	_, err := LoadGateway(configPath)
	if err == nil {
		t.Fatalf("expected error for invalid toml")
	}
	if !strings.Contains(err.Error(), "failed to decode config") {
		t.Fatalf("expected decode error, got %q", err.Error())
	}
}

func TestLoadInvalidToml(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.toml")
	configContent := `[belfast
invalid toml syntax
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatalf("expected error for invalid toml")
	}
	expectedErr := "failed to decode config"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Fatalf("expected error containing %q, got %q", expectedErr, err.Error())
	}
}

func TestPersistMaintenance(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	configContent := `[belfast]
bind_address = "127.0.0.1"
port = 8080
maintenance = false
name = "Another Server"

[api]
enabled = true

[database]
path = "test.db"

[region]
default = "JP"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if err := cfg.PersistMaintenance(true); err != nil {
		t.Fatalf("failed to persist maintenance: %v", err)
	}

	updatedCfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}

	if updatedCfg.Belfast.Maintenance != true {
		t.Fatalf("expected maintenance true after persist, got %v", updatedCfg.Belfast.Maintenance)
	}
}

func TestPersistMaintenanceToFalse(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	configContent := `[belfast]
bind_address = "127.0.0.1"
port = 8080
maintenance = true
name = "Another Server"

[api]
enabled = true

[database]
path = "test.db"

[region]
default = "KR"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if err := cfg.PersistMaintenance(false); err != nil {
		t.Fatalf("failed to persist maintenance: %v", err)
	}

	updatedCfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}

	if updatedCfg.Belfast.Maintenance != false {
		t.Fatalf("expected maintenance false after persist, got %v", updatedCfg.Belfast.Maintenance)
	}
}

func TestCurrent(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "current.toml")
	configContent := `[belfast]
bind_address = "0.0.0.0"
port = 9999

[db]
path = "current.db"

[region]
default = "TW"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	if _, err := Load(configPath); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if Current().Belfast.Port != 9999 {
		t.Fatalf("expected Current() to return loaded config with port 9999, got %d", Current().Belfast.Port)
	}
}

func TestSplitKeyValue(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantKey string
		wantVal string
		wantOK  bool
	}{
		{
			name:    "valid key=value",
			line:    "key = value",
			wantKey: "key",
			wantVal: "value",
			wantOK:  true,
		},
		{
			name:    "spaces around key and value",
			line:    "  key  =  value  ",
			wantKey: "key",
			wantVal: "value",
			wantOK:  true,
		},
		{
			name:    "empty value",
			line:    "key = ",
			wantKey: "key",
			wantVal: "",
			wantOK:  false,
		},
		{
			name:    "missing equals sign",
			line:    "key value",
			wantKey: "",
			wantVal: "",
			wantOK:  false,
		},
		{
			name:    "empty key",
			line:    " = value",
			wantKey: "",
			wantVal: "",
			wantOK:  false,
		},
		{
			name:    "empty value after equals",
			line:    "key =",
			wantKey: "",
			wantVal: "",
			wantOK:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKey, gotVal, gotOK := splitKeyValue(tt.line)
			if gotOK != tt.wantOK {
				t.Fatalf("expected ok %v, got %v", tt.wantOK, gotOK)
			}
			if gotOK {
				if gotKey != tt.wantKey {
					t.Fatalf("expected key %q, got %q", tt.wantKey, gotKey)
				}
				if gotVal != tt.wantVal {
					t.Fatalf("expected value %q, got %q", tt.wantVal, gotVal)
				}
			}
		})
	}
}

func TestInsertMaintenanceFlag(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		enabled  bool
		expected []string
	}{
		{
			name:     "insert at beginning of belfast section",
			lines:    []string{"[other]", "key = value"},
			enabled:  true,
			expected: []string{"[other]", "key = value", "", "[belfast]", "maintenance = true"},
		},
		{
			name:     "insert at end of belfast section",
			lines:    []string{"[other]", "[belfast]", "key = value", "[another]"},
			enabled:  false,
			expected: []string{"[other]", "[belfast]", "maintenance = false", "key = value", "[another]"},
		},
		{
			name:     "empty file",
			lines:    []string{},
			enabled:  true,
			expected: []string{"", "[belfast]", "maintenance = true"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := insertMaintenanceFlag(tt.lines, tt.enabled)
			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d lines, got %d", len(tt.expected), len(result))
			}
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Fatalf("expected line %d to be %q, got %q", i, expected, result[i])
				}
			}
		})
	}
}

func TestUpdateMaintenanceFlagAddsWhenMissing(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	configContent := `[belfast]
bind_address = "127.0.0.1"
port = 8080

[api]
enabled = true
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	if err := updateMaintenanceFlag(configPath, true); err != nil {
		t.Fatalf("update maintenance: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config file: %v", err)
	}

	if !strings.Contains(string(data), "maintenance = true") {
		t.Fatalf("expected maintenance line to be inserted")
	}
}

func TestUpdateMaintenanceFlagMissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "missing.toml")

	if err := updateMaintenanceFlag(configPath, false); err == nil {
		t.Fatalf("expected error for missing config file")
	}
}

func TestUpdateMaintenanceFlagStatError(t *testing.T) {
	oldRead := readFile
	oldStat := statFile
	oldWrite := writeFile
	t.Cleanup(func() {
		readFile = oldRead
		statFile = oldStat
		writeFile = oldWrite
	})

	readFile = os.ReadFile
	writeFile = os.WriteFile
	statFile = func(string) (os.FileInfo, error) {
		return nil, errors.New("stat failed")
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	configContent := `[belfast]
maintenance = false
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	if err := updateMaintenanceFlag(configPath, true); err == nil {
		t.Fatalf("expected stat error")
	}
}

func TestUpdateMaintenanceFlagReadError(t *testing.T) {
	oldRead := readFile
	oldStat := statFile
	oldWrite := writeFile
	t.Cleanup(func() {
		readFile = oldRead
		statFile = oldStat
		writeFile = oldWrite
	})

	readFile = func(string) ([]byte, error) {
		return nil, errors.New("read failed")
	}
	statFile = os.Stat
	writeFile = os.WriteFile

	if err := updateMaintenanceFlag("config.toml", true); err == nil {
		t.Fatalf("expected read error")
	}
}

func TestUpdateMaintenanceFlagWriteError(t *testing.T) {
	oldRead := readFile
	oldStat := statFile
	oldWrite := writeFile
	t.Cleanup(func() {
		readFile = oldRead
		statFile = oldStat
		writeFile = oldWrite
	})

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	configContent := `[belfast]
maintenance = false
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	readFile = os.ReadFile
	statFile = os.Stat
	writeFile = func(string, []byte, os.FileMode) error {
		return errors.New("write failed")
	}

	if err := updateMaintenanceFlag(configPath, true); err == nil {
		t.Fatalf("expected write error")
	}
}

func TestUpdateMaintenanceFlagUpdatesExisting(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	configContent := `[belfast]
maintenance = false
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	if err := updateMaintenanceFlag(configPath, true); err != nil {
		t.Fatalf("update maintenance: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "maintenance = true") {
		t.Fatalf("expected maintenance to be updated to true")
	}
	if strings.Contains(content, "maintenance = false") {
		t.Fatalf("expected maintenance false to be replaced")
	}
}

func TestApplySchemaName(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		schemaName string
		expected   string
	}{
		{
			name:       "empty schema name",
			path:       "data/belfast.db",
			schemaName: "",
			expected:   "data/belfast.db",
		},
		{
			name:       "replace base with schema",
			path:       "data/belfast.db",
			schemaName: "custom",
			expected:   filepath.Join("data", "custom.db"),
		},
		{
			name:       "no extension",
			path:       "belfast",
			schemaName: "custom",
			expected:   "custom.db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applySchemaName(tt.path, tt.schemaName)
			if got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "whitespace only",
			input:    "   ",
			expected: "",
		},
		{
			name:     "letters spaces and digits",
			input:    " My Server 2 ",
			expected: "my_server_2",
		},
		{
			name:     "punctuation only",
			input:    "!@#",
			expected: "",
		},
		{
			name:     "already snake",
			input:    "Already_Snake",
			expected: "already_snake",
		},
		{
			name:     "trims trailing underscore",
			input:    "Hello!",
			expected: "hello",
		},
		{
			name:     "collapses multiple separators",
			input:    "Hello---World",
			expected: "hello_world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toSnakeCase(tt.input)
			if got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
