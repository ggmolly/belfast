package answer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/misc"
	"github.com/ggmolly/belfast/internal/protobuf"
	"github.com/ggmolly/belfast/internal/region"
	"google.golang.org/protobuf/proto"
)

func VersionCheck(buffer *[]byte, client *connection.Client) (int, int, error) {
	const responseID = 10997
	var versionCheck protobuf.CS_10996
	err := proto.Unmarshal(*buffer, &versionCheck)
	if err != nil {
		return 0, responseID, err
	}

	belfastRegion := region.Current()
	versionString, err := misc.ResolveRegionVersion(belfastRegion)
	if err != nil {
		return 0, responseID, err
	}

	versionParts, err := parseVersionParts(versionString)
	if err != nil {
		return 0, responseID, err
	}

	url, ok := consts.GamePlatformUrl[belfastRegion][versionCheck.GetPlatform()]
	if !ok {
		resolvedPlatform, ok := platformMap[versionCheck.GetPlatform()]
		if !ok {
			resolvedPlatform = "Unknown"
		}
		return 0, responseID, fmt.Errorf("unknown platform '%s' (id='%s')", resolvedPlatform, versionCheck.GetPlatform())
	}

	response := protobuf.SC_10997{
		Version1:    proto.Uint32(versionParts[0]),
		Version2:    proto.Uint32(versionParts[1]),
		Version3:    proto.Uint32(versionParts[2]),
		Version4:    proto.Uint32(versionParts[3]),
		GatewayIp:   proto.String(consts.RegionGateways[belfastRegion]),
		GatewayPort: proto.Uint32(80),
		Url:         proto.String(url),
	}

	return client.SendMessage(responseID, &response)
}

func parseVersionParts(version string) ([4]uint32, error) {
	var parts [4]uint32
	if version == "" {
		return parts, fmt.Errorf("empty version")
	}

	segments := strings.Split(version, ".")
	if len(segments) < 3 || len(segments) > 4 {
		return parts, fmt.Errorf("invalid version format %q", version)
	}

	for i, segment := range segments {
		value, err := strconv.ParseUint(segment, 10, 32)
		if err != nil {
			return parts, fmt.Errorf("invalid version segment %q in %q", segment, version)
		}
		parts[i] = uint32(value)
	}

	return parts, nil
}
