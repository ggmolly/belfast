package misc

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

var (
	localVersionsOnce sync.Once
	localVersions     map[string]string
	localVersionsErr  error
)

func ResolveRegionVersion(region string) (string, error) {
	versions := GetLatestVersions()
	if versions != nil {
		if entry, ok := versions[region]; ok {
			return entry.Version, nil
		}
	}

	local, err := loadLocalVersions()
	if err != nil {
		return "", err
	}

	version, ok := local[region]
	if !ok {
		return "", fmt.Errorf("missing version for region %q", region)
	}

	return version, nil
}

func loadLocalVersions() (map[string]string, error) {
	localVersionsOnce.Do(func() {
		file, err := os.Open("belfast-data/versions.json")
		if err != nil {
			localVersionsErr = err
			return
		}
		defer file.Close()
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&localVersions); err != nil {
			localVersionsErr = err
		}
	})

	return localVersions, localVersionsErr
}
