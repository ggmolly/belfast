package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Belfast BelfastConfig  `toml:"belfast"`
	API     APIConfig      `toml:"api"`
	DB      DatabaseConfig `toml:"database"`
	Region  RegionConfig   `toml:"region"`
}

type BelfastConfig struct {
	BindAddress string `toml:"bind_address"`
	Port        int    `toml:"port"`
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

func Load(path string) (Config, error) {
	var cfg Config
	if _, err := os.Stat(path); err != nil {
		return cfg, fmt.Errorf("config file missing: %w", err)
	}
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to decode config: %w", err)
	}
	return cfg, nil
}
