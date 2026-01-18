package api

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Enabled     bool
	Env         string
	Port        int
	CORSOrigins []string
}

func LoadConfig() Config {
	cfg := Config{
		Enabled:     envBool("API_ENABLED", true),
		Env:         envString("API_ENV", "development"),
		Port:        envInt("API_PORT", 6669),
		CORSOrigins: envCSV("API_CORS_ORIGINS"),
	}
	if cfg.Env == "development" && len(cfg.CORSOrigins) == 0 {
		cfg.CORSOrigins = []string{"*"}
	}
	return cfg
}

func envString(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envCSV(key string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	output := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		output = append(output, trimmed)
	}
	return output
}
