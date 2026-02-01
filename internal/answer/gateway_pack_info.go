package answer

import (
	"fmt"
	"time"

	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/misc"
	"github.com/ggmolly/belfast/internal/protobuf"
	"github.com/ggmolly/belfast/internal/region"
	"google.golang.org/protobuf/proto"
)

func GatewayPackInfo(buffer *[]byte, client *connection.Client) (int, int, error) {
	const packetId = 10701
	var payload protobuf.CS_10700
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, packetId, err
	}
	updateVersions(misc.GetGameHashes)
	belfastRegion := region.Current()
	resolvedPlatform, ok := platformMap[payload.GetPlatform()]
	if !ok {
		resolvedPlatform = "Unknown"
	}
	url, ok := consts.GamePlatformUrl[belfastRegion][payload.GetPlatform()]
	if !ok {
		return 0, packetId, fmt.Errorf("unknown platform '%s' (id='%s')", resolvedPlatform, payload.GetPlatform())
	}
	response := protobuf.SC_10701{
		Url:                     proto.String(url),
		Version:                 versions,
		AddrList:                buildGatewayAddrList(config.Current().Servers),
		Timestamp:               proto.Uint32(uint32(time.Now().Unix())),
		Monday_0OclockTimestamp: proto.Uint32(consts.Monday_0OclockTimestamps[belfastRegion]),
		CdnList:                 []string{},
	}
	return client.SendMessage(packetId, &response)
}

func buildGatewayAddrList(servers []config.ServerConfig) []*protobuf.LOGIN_ADDR {
	statuses := getServerStatusCache(servers)
	output := make([]*protobuf.LOGIN_ADDR, 0, len(servers))
	for _, server := range servers {
		status, ok := statuses[server.ID]
		desc := server.IP
		if ok {
			if status.Name != "" {
				desc = status.Name
			}
			desc = formatServerName(desc, status.Commit)
		} else {
			desc = formatServerName(desc, "")
		}
		proxyIP := ""
		if server.ProxyIP != nil {
			proxyIP = *server.ProxyIP
		}
		proxyPort := uint32(0)
		if server.ProxyPort != nil && *server.ProxyPort > 0 {
			proxyPort = uint32(*server.ProxyPort)
		}
		addr := &protobuf.LOGIN_ADDR{
			Desc:      proto.String(desc),
			Ip:        proto.String(server.IP),
			Port:      proto.Uint32(server.Port),
			ProxyIp:   proto.String(proxyIP),
			ProxyPort: proto.Uint32(proxyPort),
			Type:      proto.Uint32(0),
		}
		output = append(output, addr)
	}
	return output
}
