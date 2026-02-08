package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Belfast      BelfastConfig      `toml:"belfast"`
	API          APIConfig          `toml:"api"`
	Auth         AuthConfig         `toml:"auth"`
	UserAuth     AuthConfig         `toml:"user_auth"`
	DB           DatabaseConfig     `toml:"database"`
	Region       RegionConfig       `toml:"region"`
	CreatePlayer CreatePlayerConfig `toml:"create_player"`
	Servers      []ServerConfig     `toml:"servers"`
	Path         string             `toml:"-"`
}

type GatewayConfig struct {
	BindAddress string `toml:"bind_address"`
	Port        int    `toml:"port"`
	Mode        string `toml:"mode"`
	ProxyRemote string `toml:"proxy_remote"`
	// Timeout (in ms) when dialing proxy_remote in gateway proxy mode.
	ProxyDialTimeoutMS int `toml:"proxy_dial_timeout_ms"`
	// When nil, defaults to true.
	RequirePrivateClients *bool          `toml:"require_private_clients"`
	Servers               []ServerConfig `toml:"servers"`
	Path                  string         `toml:"-"`
}

type BelfastConfig struct {
	BindAddress string `toml:"bind_address"`
	Port        int    `toml:"port"`
	Maintenance bool   `toml:"maintenance"`
	Name        string `toml:"name"`
}

type ServerConfig struct {
	ID           uint32  `toml:"id"`
	IP           string  `toml:"ip"`
	Port         uint32  `toml:"port"`
	ApiPort      int     `toml:"api_port"`
	AssertOnline bool    `toml:"assert_online"`
	ProxyIP      *string `toml:"proxy_ip"`
	ProxyPort    *int    `toml:"proxy_port"`
}

type APIConfig struct {
	Enabled     bool     `toml:"enabled"`
	Port        int      `toml:"port"`
	Environment string   `toml:"environment"`
	CORSOrigins []string `toml:"cors_origins"`
}

type AuthConfig struct {
	DisableAuth                 bool         `toml:"disable_auth"`
	SessionTTLSeconds           int          `toml:"session_ttl_seconds"`
	SessionSliding              bool         `toml:"session_sliding"`
	CookieName                  string       `toml:"cookie_name"`
	CookieSecure                bool         `toml:"cookie_secure"`
	CookieSameSite              string       `toml:"cookie_same_site"`
	CSRFTokenTTLSeconds         int          `toml:"csrf_ttl_seconds"`
	PasswordMinLength           int          `toml:"password_min_length"`
	PasswordMaxLength           int          `toml:"password_max_length"`
	PasswordHashParams          Argon2Config `toml:"password_hash_params"`
	WebAuthnRPID                string       `toml:"webauthn_rp_id"`
	WebAuthnRPName              string       `toml:"webauthn_rp_name"`
	WebAuthnExpectedOrigins     []string     `toml:"webauthn_expected_origins"`
	WebAuthnChallengeTTLSeconds int          `toml:"webauthn_challenge_ttl_seconds"`
	RateLimitWindowSeconds      int          `toml:"rate_limit_window_seconds"`
	RateLimitLoginMax           int          `toml:"rate_limit_login_max"`
	RateLimitPasskeyMax         int          `toml:"rate_limit_passkey_max"`
}

type Argon2Config struct {
	Memory      uint32 `toml:"memory"`
	Iterations  uint32 `toml:"iterations"`
	Parallelism uint8  `toml:"parallelism"`
	SaltLength  uint32 `toml:"salt_length"`
	KeyLength   uint32 `toml:"key_length"`
}

type DatabaseConfig struct {
	Driver     string `toml:"driver"`
	Path       string `toml:"path"`
	DSN        string `toml:"dsn"`
	SchemaName string `toml:"schema_name"`
}

type RegionConfig struct {
	Default string `toml:"default"`
}

type CreatePlayerConfig struct {
	SkipOnboarding     bool     `toml:"skip_onboarding"`
	NameBlacklist      []string `toml:"name_blacklist"`
	NameIllegalPattern string   `toml:"name_illegal_pattern"`
}

var current Config

var (
	readFile  = os.ReadFile
	statFile  = os.Stat
	writeFile = os.WriteFile
)

func Current() Config {
	return current
}

func Load(path string) (Config, error) {
	var cfg Config
	if _, err := os.Stat(path); err != nil {
		return cfg, fmt.Errorf("config file missing: %w", err)
	}
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to decode config: %w", err)
	}
	cfg.DB.Driver = normalizeDBDriver(cfg.DB.Driver)
	if cfg.Belfast.Port == 0 {
		cfg.Belfast.Port = 80
	}
	schemaName := resolveSchemaName(cfg)
	if schemaName != "" && cfg.DB.SchemaName == "" {
		cfg.DB.SchemaName = schemaName
	}
	if cfg.DB.Driver == "sqlite" {
		if cfg.DB.Path == "" {
			cfg.DB.Path = "data/belfast.db"
		}
		if schemaName != "" {
			cfg.DB.Path = applySchemaName(cfg.DB.Path, schemaName)
		}
	}
	cfg.Path = path
	current = cfg
	return cfg, nil
}

func normalizeDBDriver(driver string) string {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "", "sqlite", "sqlite3":
		return "sqlite"
	case "postgres", "postgresql", "pg":
		return "postgres"
	case "mysql":
		return "mysql"
	default:
		return strings.ToLower(strings.TrimSpace(driver))
	}
}

func LoadGateway(path string) (GatewayConfig, error) {
	var cfg GatewayConfig
	if _, err := os.Stat(path); err != nil {
		return cfg, fmt.Errorf("config file missing: %w", err)
	}
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to decode config: %w", err)
	}
	if cfg.Port == 0 {
		cfg.Port = 80
	}
	if strings.TrimSpace(cfg.Mode) == "" {
		cfg.Mode = "serve"
	}
	if cfg.ProxyDialTimeoutMS == 0 {
		cfg.ProxyDialTimeoutMS = 5000
	}
	if cfg.RequirePrivateClients == nil {
		defaultRequirePrivate := true
		cfg.RequirePrivateClients = &defaultRequirePrivate
	}
	cfg.Path = path
	current = Config{
		Belfast: BelfastConfig{
			BindAddress: cfg.BindAddress,
			Port:        cfg.Port,
		},
		Servers: cfg.Servers,
		Path:    cfg.Path,
	}
	return cfg, nil
}

func (cfg *Config) PersistMaintenance(enabled bool) error {
	cfg.Belfast.Maintenance = enabled
	return updateMaintenanceFlag(cfg.Path, enabled)
}

func updateMaintenanceFlag(path string, enabled bool) error {
	data, err := readFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	section := ""
	updated := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			section = strings.TrimSpace(trimmed[1 : len(trimmed)-1])
			continue
		}
		if section != "belfast" {
			continue
		}
		key, _, ok := splitKeyValue(trimmed)
		if ok && key == "maintenance" {
			lines[i] = fmt.Sprintf("maintenance = %t", enabled)
			updated = true
			break
		}
	}
	if !updated {
		lines = insertMaintenanceFlag(lines, enabled)
	}
	info, err := statFile(path)
	if err != nil {
		return err
	}
	return writeFile(path, []byte(strings.Join(lines, "\n")), info.Mode().Perm())
}

func insertMaintenanceFlag(lines []string, enabled bool) []string {
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "[belfast]" {
			insert := fmt.Sprintf("maintenance = %t", enabled)
			return append(lines[:i+1], append([]string{insert}, lines[i+1:]...)...)
		}
	}
	insert := []string{"", "[belfast]", fmt.Sprintf("maintenance = %t", enabled)}
	return append(lines, insert...)
}

func splitKeyValue(line string) (string, string, bool) {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	if key == "" || value == "" {
		return "", "", false
	}
	return key, value, true
}

func resolveSchemaName(cfg Config) string {
	if cfg.DB.SchemaName != "" {
		return cfg.DB.SchemaName
	}
	if cfg.Belfast.Name == "" {
		return ""
	}
	return toSnakeCase(cfg.Belfast.Name)
}

func applySchemaName(path string, schemaName string) string {
	if schemaName == "" {
		return path
	}
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	if ext == "" {
		ext = ".db"
	}
	name := schemaName + ext
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return name
	}
	return filepath.Join(dir, name)
}

func toSnakeCase(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	var output []rune
	underscore := false
	for _, r := range trimmed {
		if r >= 'A' && r <= 'Z' {
			r = r + ('a' - 'A')
		}
		isLower := r >= 'a' && r <= 'z'
		isDigit := r >= '0' && r <= '9'
		if isLower || isDigit {
			output = append(output, r)
			underscore = false
			continue
		}
		if !underscore && len(output) > 0 {
			output = append(output, '_')
			underscore = true
		}
	}
	if len(output) == 0 {
		return ""
	}
	if output[len(output)-1] == '_' {
		output = output[:len(output)-1]
	}
	return string(output)
}
