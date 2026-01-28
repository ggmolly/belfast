package region

import (
	"fmt"
	"os"
)

var currentRegion string

func Current() string {
	if currentRegion != "" {
		return currentRegion
	}
	value := os.Getenv("AL_REGION")
	if value == "" {
		return "EN"
	}
	return value
}

func SetCurrent(region string) error {
	if err := Validate(region); err != nil {
		return err
	}
	currentRegion = region
	return nil
}

func Validate(region string) error {
	switch region {
	case "CN", "EN", "JP", "KR", "TW":
		return nil
	default:
		return fmt.Errorf("invalid region %q", region)
	}
}

func ResetCurrentForTest() {
	currentRegion = ""
}
