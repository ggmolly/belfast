package misc

import "os"

// GetSpecifiedRegion returns the value of the environment variable AL_REGION
// it is needed for the web UI to highlight the correct region since we cannot call os.Getenv
// from the web template engine ¯\_(ツ)_/¯
func GetSpecifiedRegion() string {
	return os.Getenv("AL_REGION")
}
