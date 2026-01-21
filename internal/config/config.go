package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Belfast      BelfastConfig      `toml:"belfast"`
	API          APIConfig          `toml:"api"`
	DB           DatabaseConfig     `toml:"database"`
	Region       RegionConfig       `toml:"region"`
	CreatePlayer CreatePlayerConfig `toml:"create_player"`
	Path         string             `toml:"-"`
}

type BelfastConfig struct {
	BindAddress string `toml:"bind_address"`
	Port        int    `toml:"port"`
	Maintenance bool   `toml:"maintenance"`
	ServerHost  string `toml:"server_host"`
}

type APIConfig struct {
	Enabled     bool     `toml:"enabled"`
	Port        int      `toml:"port"`
	Environment string   `toml:"environment"`
	CORSOrigins []string `toml:"cors_origins"`
}

type DatabaseConfig struct {
	Path string `toml:"path"`
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
	cfg.Path = path
	current = cfg
	return cfg, nil
}

func (cfg *Config) PersistMaintenance(enabled bool) error {
	cfg.Belfast.Maintenance = enabled
	return updateMaintenanceFlag(cfg.Path, enabled)
}

func updateMaintenanceFlag(path string, enabled bool) error {
	data, err := os.ReadFile(path)
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
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), info.Mode().Perm())
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
