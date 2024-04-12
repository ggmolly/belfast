package answer

import (
	"fmt"
	"time"

	"github.com/ggmolly/belfast/connection"
	"github.com/ggmolly/belfast/misc"

	"github.com/ggmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

var (
	platformMap = map[string]string{
		"0": "Android", // supposedly
		"1": "iOS",     // known
		// maybe more?
	}
	versions []string
)

// Answer to a CS_10800 packet with a SC_10801 packet + hashes
func Forge_SC10801(buffer *[]byte, client *connection.Client) (int, int, error) {
	const packetId = 10801
	var updateCheck protobuf.CS_10800
	err := proto.Unmarshal(*buffer, &updateCheck)
	if err != nil {
		return 0, packetId, err
	}

	if len(versions) == 0 {
		hashes := misc.GetGameHashes()
		for _, hash := range hashes {
			versions = append(versions, hash.Hash)
		}
		versions = append(versions, "dTag-1")
	}

	response := protobuf.SC_10801{
		GatewayIp:               proto.String("blhxusgate.yo-star.com"),
		GatewayPort:             proto.Uint32(80),
		Url:                     proto.String(""),
		Version:                 versions,
		ProxyIp:                 proto.String("blhxusproxy.yo-star.com"),
		ProxyPort:               proto.Uint32(20000),
		IsTs:                    proto.Uint32(0),
		Timestamp:               proto.Uint32(uint32(time.Now().Unix())),
		Monday_0OclockTimestamp: proto.Uint32(1606114800), // 23/11/2020 08:00:00

		// wtf is this i don't even understand what monday_0oclock_timestamp is
		// who would even do such a thing
	}

	resolvedPlatform, ok := platformMap[updateCheck.GetPlatform()]
	if !ok {
		resolvedPlatform = "Unknown"
	}

	if updateCheck.GetPlatform() == "1" { // iOS, set the iTunes URL
		response.Url = proto.String("https://itunes.apple.com/us/app/azur-lane/id1411126549")
	} else if updateCheck.GetPlatform() == "0" { // Android, set the Play Store URL (untested)
		response.Url = proto.String("https://play.google.com/store/apps/details?id=com.YoStarEN.AzurLane")
	} else { // Unsupported platform
		return 0, 10801, fmt.Errorf("unknown platform '%s' (id='%s')", resolvedPlatform, updateCheck.GetPlatform())
	}

	return client.SendMessage(packetId, &response)
}
