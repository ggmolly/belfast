package region

import "os"

func Current() string {
	value := os.Getenv("AL_REGION")
	if value == "" {
		return "EN"
	}
	return value
}
