package api

import (
	"strings"

	"github.com/ggmolly/belfast/internal/config"
)

type Config struct {
	Enabled       bool
	Env           string
	Port          int
	CORSOrigins   []string
	RuntimeConfig *config.Config
}

func LoadConfig(cfg config.Config) Config {
	apiConfig := Config{
		Enabled:       cfg.API.Enabled,
		Env:           cfg.API.Environment,
		Port:          cfg.API.Port,
		CORSOrigins:   cfg.API.CORSOrigins,
		RuntimeConfig: &cfg,
	}
	if apiConfig.Env == "development" && len(apiConfig.CORSOrigins) == 0 {
		apiConfig.CORSOrigins = []string{"*"}
	}
	apiConfig.CORSOrigins = normalizeOrigins(apiConfig.CORSOrigins)
	return apiConfig
}

func normalizeOrigins(origins []string) []string {
	output := make([]string, 0, len(origins))
	for _, origin := range origins {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" {
			continue
		}
		output = append(output, trimmed)
	}
	return output
}
