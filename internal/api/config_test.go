package api

import (
	"reflect"
	"testing"

	"github.com/ggmolly/belfast/internal/config"
)

func TestNormalizeOrigins(t *testing.T) {
	input := []string{" http://a.com ", "", " ", "http://b.com"}
	result := normalizeOrigins(input)
	if !reflect.DeepEqual(result, []string{"http://a.com", "http://b.com"}) {
		t.Fatalf("unexpected origins: %v", result)
	}
}

func TestLoadConfigDevelopmentDefaults(t *testing.T) {
	base := config.Config{
		API: config.APIConfig{
			Enabled:     true,
			Port:        8080,
			Environment: "development",
			CORSOrigins: nil,
		},
	}
	loaded := LoadConfig(base)
	if !loaded.Enabled {
		t.Fatalf("expected enabled config")
	}
	if len(loaded.CORSOrigins) != 1 || loaded.CORSOrigins[0] != "*" {
		t.Fatalf("expected default CORS origins, got %v", loaded.CORSOrigins)
	}
}

func TestLoadConfigNormalizesOrigins(t *testing.T) {
	base := config.Config{
		API: config.APIConfig{
			Enabled:     true,
			Port:        8080,
			Environment: "production",
			CORSOrigins: []string{" http://a.com ", ""},
		},
	}
	loaded := LoadConfig(base)
	if len(loaded.CORSOrigins) != 1 || loaded.CORSOrigins[0] != "http://a.com" {
		t.Fatalf("unexpected normalized origins: %v", loaded.CORSOrigins)
	}
}
