package answer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

const (
	SERVER_STATE_ONLINE  = 0
	SERVER_STATE_OFFLINE = 1
	SERVER_STATE_FULL    = 2
	SERVER_STATE_BUSY    = 3
)

var (
	// Server list
	Servers = []*protobuf.SERVERINFO{}
)

func buildServerInfo(servers []config.ServerConfig, statuses map[uint32]serverStatusEntry) []*protobuf.SERVERINFO {
	output := make([]*protobuf.SERVERINFO, len(servers))
	for i, server := range servers {
		status, ok := statuses[server.ID]
		state := uint32(SERVER_STATE_OFFLINE)
		name := server.IP
		if ok {
			if status.Name != "" {
				name = status.Name
			}
			state = status.State
			name = formatServerName(name, status.Commit)
		} else {
			name = formatServerName(name, "")
		}
		proxyIP := ""
		if server.ProxyIP != nil {
			proxyIP = *server.ProxyIP
		}
		proxyPort := uint32(0)
		if server.ProxyPort != nil && *server.ProxyPort > 0 {
			proxyPort = uint32(*server.ProxyPort)
		}
		info := &protobuf.SERVERINFO{
			Ids:       []uint32{server.ID},
			Ip:        proto.String(server.IP),
			Port:      proto.Uint32(server.Port),
			State:     proto.Uint32(state),
			Name:      proto.String(name),
			TagState:  proto.Uint32(0),
			Sort:      proto.Uint32(uint32(i + 1)),
			ProxyIp:   proto.String(proxyIP),
			ProxyPort: proto.Uint32(proxyPort),
		}
		output[i] = info
	}
	return output
}

func formatServerName(name string, commit string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		trimmed = "Unknown"
	}
	if commit == "" {
		return trimmed
	}
	return fmt.Sprintf("%s (%s)", trimmed, commit)
}

// Answer to a pseudo CS_8239 packet with a SC_8239 packet + server list (HTTP/1.1 200 OK)
func Forge_SC8239(buffer *[]byte, client *connection.Client) (int, int, error) {
	const packetId = 8239
	var answerBuffer bytes.Buffer

	servers := config.Current().Servers
	statuses := getServerStatusCache(servers)
	Servers = buildServerInfo(servers, statuses)

	// Write the HTTP header
	answerBuffer.WriteString("HTTP/1.1 200 OK\r\nContent-Type: text/plain;charset=utf-8\r\nAccess-Control-Allow-Origin: \r\nContent-Length: 335\r\n\r\n")

	// Write the JSON-ized server list
	jsonData, err := json.Marshal(Servers)
	if err != nil {
		return 0, packetId, err
	}
	answerBuffer.Write(jsonData)

	n, err := (*client.Connection).Write(answerBuffer.Bytes())
	if err != nil {
		return 0, packetId, err
	}
	return n, packetId, nil
}
