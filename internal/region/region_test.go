package region

import (
	"testing"
)

func TestValidateValidRegions(t *testing.T) {
	validRegions := []string{"CN", "EN", "JP", "KR", "TW"}

	for _, region := range validRegions {
		t.Run("valid_"+region, func(t *testing.T) {
			if err := Validate(region); err != nil {
				t.Fatalf("expected no error for region %q, got %v", region, err)
			}
		})
	}
}

func TestValidateInvalidRegion(t *testing.T) {
	invalidRegions := []string{"cn", "en", "jp", "kr", "tw", "", "US", "EU", "cn-en", "en-us"}

	for _, region := range invalidRegions {
		t.Run("invalid_"+region, func(t *testing.T) {
			if err := Validate(region); err == nil {
				t.Fatalf("expected error for invalid region %q", region)
			}
		})
	}
}

func TestSetCurrentValidRegion(t *testing.T) {
	originalEnv := ""
	t.Setenv("AL_REGION", originalEnv)
	t.Cleanup(func() {
		t.Setenv("AL_REGION", originalEnv)
	})

	validRegions := []string{"CN", "EN", "JP", "KR", "TW"}
	for _, region := range validRegions {
		err := SetCurrent(region)
		if err != nil {
			t.Fatalf("expected no error for valid region %q, got %v", region, err)
		}
		if Current() != region {
			t.Fatalf("expected Current() to return %q after SetCurrent(), got %q", region, Current())
		}
	}
}

func TestSetCurrentInvalidRegion(t *testing.T) {
	originalEnv := ""
	t.Setenv("AL_REGION", originalEnv)
	t.Cleanup(func() {
		t.Setenv("AL_REGION", originalEnv)
	})

	invalidRegions := []string{"XX", "YY", "ZZ"}
	for _, region := range invalidRegions {
		err := SetCurrent(region)
		if err == nil {
			t.Fatalf("expected error for invalid region %q", region)
		}
	}
}

func TestCurrentReturnsEnv(t *testing.T) {
	currentRegion = ""

	originalEnv := ""
	t.Setenv("AL_REGION", originalEnv)
	t.Cleanup(func() {
		t.Setenv("AL_REGION", originalEnv)
		currentRegion = ""
	})

	envRegions := []string{"", "CN", "JP", "KR", "TW"}
	for _, envRegion := range envRegions {
		t.Setenv("AL_REGION", envRegion)

		expected := envRegion
		if expected == "" {
			expected = "EN"
		}

		result := Current()
		if result != expected {
			t.Fatalf("expected Current() to return %q when AL_REGION=%q, got %q", expected, envRegion, result)
		}
	}
}

func TestCurrentReturnsCached(t *testing.T) {
	currentRegion = ""

	originalEnv := ""
	t.Setenv("AL_REGION", originalEnv)
	t.Cleanup(func() {
		t.Setenv("AL_REGION", originalEnv)
		currentRegion = ""
	})

	if err := SetCurrent("JP"); err != nil {
		t.Fatalf("failed to set current region: %v", err)
	}
	if Current() != "JP" {
		t.Fatalf("expected Current() to return 'JP', got %s", Current())
	}

	t.Setenv("AL_REGION", "")
	if Current() != "JP" {
		t.Fatalf("expected Current() to return cached 'JP' after env is cleared, got %s", Current())
	}
}

func TestResetCurrentForTest(t *testing.T) {
	if err := SetCurrent("KR"); err != nil {
		t.Fatalf("failed to set current region: %v", err)
	}
	ResetCurrentForTest()
	if Current() == "KR" {
		t.Fatalf("expected reset to clear cached region")
	}
}
