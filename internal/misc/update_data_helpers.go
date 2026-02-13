package misc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ggmolly/belfast/internal/logger"
)

const (
	// region / file
	URL_BASE            = "https://raw.githubusercontent.com/ggmolly/belfast-data/main/%s/%s"
	REGIONLESS_URL_BASE = "https://raw.githubusercontent.com/ggmolly/belfast-data/main/%s"
)

var (
	// Golang maps are unordered, so we keep track of key order ourselves.
	order = []string{"Items", "Buffs", "Ships", "Skins", "Resources", "Pools", "Requisition", "BuildTimes", "ShopOffers", "Weapons", "Equipments", "Skills", "Configs", "JuustagramTemplates", "JuustagramNpcTemplates", "JuustagramLanguage", "JuustagramShipGroups"}
)

func getBelfastData(region string, file string) (*json.Decoder, *http.Response, error) {
	var url string
	if region == "" {
		url = fmt.Sprintf(REGIONLESS_URL_BASE, file)
	} else {
		url = fmt.Sprintf(URL_BASE, region, file)
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("failed to fetch data from %s: %s", url, resp.Status)
	}
	return json.NewDecoder(resp.Body), resp, nil
}

type githubContent struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func listBelfastDataFiles(region string, directory string) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/ggmolly/belfast-data/contents/%s/%s", region, directory)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list data from %s: %s", url, resp.Status)
	}
	var contents []githubContent
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&contents); err != nil {
		return nil, err
	}
	files := make([]string, 0, len(contents))
	for _, entry := range contents {
		if entry.Type != "file" || !strings.HasSuffix(entry.Name, ".json") {
			continue
		}
		files = append(files, fmt.Sprintf("%s/%s", directory, entry.Name))
	}
	return files, nil
}

func filterConfigFiles(files []string, prefixes []string, include []string) []string {
	allowed := make(map[string]struct{}, len(files))
	for _, file := range files {
		name := strings.TrimPrefix(file, "ShareCfg/")
		name = strings.TrimPrefix(name, "GameCfg/")
		for _, prefix := range prefixes {
			if strings.HasPrefix(name, prefix) {
				allowed[file] = struct{}{}
				break
			}
		}
	}
	for _, file := range include {
		allowed[file] = struct{}{}
	}
	filtered := make([]string, 0, len(allowed))
	for file := range allowed {
		filtered = append(filtered, file)
	}
	return filtered
}

func configEntryKey(raw json.RawMessage, index int) string {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return strconv.Itoa(index)
	}
	for _, field := range []string{"id", "ID", "Id"} {
		value, ok := object[field]
		if !ok {
			continue
		}
		if key, ok := parseConfigKey(value); ok {
			return key
		}
	}
	return strconv.Itoa(index)
}

func parseConfigKey(value json.RawMessage) (string, bool) {
	var number uint64
	if err := json.Unmarshal(value, &number); err == nil {
		return strconv.FormatUint(number, 10), true
	}
	var text string
	if err := json.Unmarshal(value, &text); err == nil {
		return text, text != ""
	}
	var floatValue float64
	if err := json.Unmarshal(value, &floatValue); err == nil {
		return strconv.FormatUint(uint64(floatValue), 10), true
	}
	return "", false
}

func UpdateAllData(region string) {
	logger.LogEvent("GameData", "Updating", "Updating all game data.. this may take a while.", logger.LOG_LEVEL_INFO)
	updateAllDataSQLC(region)
}
