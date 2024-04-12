package misc

import (
	"encoding/gob"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/ggmolly/belfast/connection"
	"github.com/ggmolly/belfast/logger"
	"github.com/ggmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

var (
	regex        = regexp.MustCompile(`(EN|TW|KR|CN|JP):\s(\d{1,9}\.\d{1,9}\.\d{1,9})`)
	versionFiles = map[string]string{
		"China":    "../AzurLaneLuaScripts/versions/CN.svg",
		"Japan":    "../AzurLaneLuaScripts/versions/JP.svg",
		"Korea":    "../AzurLaneLuaScripts/versions/KR.svg",
		"Taiwan":   "../AzurLaneLuaScripts/versions/TW.svg",
		"Occident": "../AzurLaneLuaScripts/versions/EN.svg",
	}

	errRegionMismatch  = errors.New("Region mismatch")
	errVersionMismatch = errors.New("Version mismatch")
)

const (
	region = "Occident"
)

type VersionMap map[string]Version
type HashMap []GameChecksum

type Version struct {
	Region  string
	Version string
}

type GameChecksum struct {
	Category string
	Hash     string
}

type hashCache struct {
	Region  string
	Version string
	Hashes  HashMap
}

var azurLaneHashes HashMap
var azurLaneVersions VersionMap

func readVersion(path string) string {
	file, err := os.Open(path)
	version := "0.0.0"
	if err != nil {
		logger.LogEvent("GameUpdate", "readVersion", err.Error(), logger.LOG_LEVEL_ERROR)
		return version
	}
	defer file.Close()
	data := make([]byte, 150)
	n, err := file.Read(data)
	if err != nil {
		logger.LogEvent("GameUpdate", "readVersion", err.Error(), logger.LOG_LEVEL_ERROR)
		return version
	}
	data = data[:n]
	// get first match
	matches := regex.FindAllStringSubmatch(string(data), -1)
	if len(matches) == 0 {
		logger.LogEvent("GameUpdate", "readVersion", fmt.Sprintf("No matches found in %s", path), logger.LOG_LEVEL_ERROR)
		return version
	}
	// get second group
	version = matches[0][2]
	return version
}

func GetLatestVersions() VersionMap {
	if azurLaneVersions != nil {
		return azurLaneVersions
	}
	exec.Command("./game_update.sh").Run()
	azurLaneVersions = make(VersionMap)
	for region, path := range versionFiles {
		azurLaneVersions[region] = Version{
			Region:  region,
			Version: readVersion(path),
		}
	}
	return azurLaneVersions
}

func hashFromCache() (HashMap, error) {
	// use gob to read the cache file
	file, err := os.Open(".cached_hashes")
	if err != nil { // no cache
		return nil, err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	var cache hashCache
	err = decoder.Decode(&cache)
	if err != nil {
		logger.LogEvent("GameUpdate", "GetHashes", err.Error(), logger.LOG_LEVEL_ERROR)
		return nil, err
	}
	if cache.Region != region {
		return nil, errRegionMismatch
	}
	if cache.Version != azurLaneVersions[region].Version {
		return nil, errVersionMismatch
	}
	return cache.Hashes, nil

}

func GetGameHashes() HashMap {
	version := GetLatestVersions()[region].Version

	if azurLaneHashes != nil && azurLaneVersions[region].Version == version {
		return azurLaneHashes
	}

	// check if we have a cached version
	hashes, err := hashFromCache()
	if err == nil && azurLaneVersions[region].Version == version {
		azurLaneHashes = hashes
		return hashes
	}

	// no cache, get the hashes from the server
	logger.LogEvent("GameUpdate", "GetHashes", "No cached hashes, fetching from server", logger.LOG_LEVEL_INFO)
	sock, err := net.Dial("tcp", "blhxusgate.yo-star.com:80")
	if err != nil {
		logger.LogEvent("GameUpdate", "GetHashes", err.Error(), logger.LOG_LEVEL_ERROR)
		return nil
	}
	defer sock.Close()

	// Forge an update packet, CS_10800
	promptUpdate := protobuf.CS_10800{
		State:    proto.Uint32(59),  // 59 is something, might need to update this later?
		Platform: proto.String("1"), // iOS
	}
	packet, err := proto.Marshal(&promptUpdate)
	if err != nil {
		logger.LogEvent("GameUpdate", "GetHashes", err.Error(), logger.LOG_LEVEL_ERROR)
		return nil
	}
	connection.InjectPacketHeader(10800, &packet, 0)
	// Send the packet
	logger.LogEvent("GameUpdate", "GetHashes", "Sending update prompt", logger.LOG_LEVEL_INFO)
	if _, err := sock.Write(packet); err != nil {
		logger.LogEvent("GameUpdate", "GetHashes", err.Error(), logger.LOG_LEVEL_ERROR)
		return nil
	}
	// Read the response
	logger.LogEvent("GameUpdate", "GetHashes", "Reading update response", logger.LOG_LEVEL_INFO)
	var responseData protobuf.SC_10801
	response := make([]byte, 1024)
	n, err := sock.Read(response)
	if err != nil || n < 8 {
		logger.LogEvent("GameUpdate", "GetHashes", "Failed to receive response, or invalid response.", logger.LOG_LEVEL_ERROR)
		return nil
	}
	response = response[7:n]
	if err := proto.Unmarshal(response, &responseData); err != nil {
		logger.LogEvent("GameUpdate", "GetHashes", err.Error(), logger.LOG_LEVEL_ERROR)
		return nil
	}
	// Parse the response
	logger.LogEvent("GameUpdate", "GetHashes", "Parsing update response", logger.LOG_LEVEL_INFO)
	for _, hash := range responseData.GetVersion() {
		if !strings.Contains(hash, "hash$") {
			continue
		}
		fields := strings.Split(hash, "$")
		azurLaneHashes = append(azurLaneHashes, GameChecksum{
			Category: fields[1],
			Hash:     hash,
		})
	}
	// Cache the hashes
	cache := hashCache{
		Region:  region,
		Version: azurLaneVersions[region].Version,
		Hashes:  azurLaneHashes,
	}
	file, err := os.OpenFile(".cached_hashes", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		logger.LogEvent("GameUpdate", "GetHashes", err.Error(), logger.LOG_LEVEL_ERROR)
		return nil
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(cache)
	if err != nil {
		logger.LogEvent("GameUpdate", "GetHashes", err.Error(), logger.LOG_LEVEL_ERROR)
		return nil
	}
	go UpdateAllData()
	return azurLaneHashes
}

func LastCacheUpdate() time.Time {
	file, err := os.Stat(".cached_hashes")
	if err != nil {
		return time.Time{}
	}
	return file.ModTime()
}

func LastCacheUpdateVersion() string {
	if azurLaneHashes == nil {
		return ""
	}
	return azurLaneVersions[region].Version
}
