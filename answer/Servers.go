package answer

import (
	"bytes"
	"encoding/json"
	"syscall"

	"github.com/bettercallmolly/belfast/connection"
	"github.com/bettercallmolly/belfast/protobuf"
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
	Servers = []*protobuf.SERVERINFO{
		{
			Ids:       []uint32{1},
			Ip:        proto.String("blhxusgs1api.yo-star.com"),
			Port:      proto.Uint32(80),
			State:     proto.Uint32(SERVER_STATE_OFFLINE),
			Name:      proto.String("Belfast - @BetterCallMolly"),
			Sort:      proto.Uint32(1),
			ProxyIp:   proto.String("blhxusproxy.yo-star.com"),
			ProxyPort: proto.Uint32(20001),
		},
		{
			Ids:       []uint32{2},
			Ip:        proto.String("blhxusgs1api.yo-star.com"),
			Port:      proto.Uint32(80),
			State:     proto.Uint32(SERVER_STATE_FULL),
			Name:      proto.String("Belfast - @BetterCallMolly"),
			Sort:      proto.Uint32(2),
			ProxyIp:   proto.String("blhxusproxy.yo-star.com"),
			ProxyPort: proto.Uint32(20001),
		},
		{
			Ids:       []uint32{3},
			Ip:        proto.String("blhxusgs1api.yo-star.com"),
			Port:      proto.Uint32(80),
			State:     proto.Uint32(SERVER_STATE_BUSY),
			Name:      proto.String("Belfast - @BetterCallMolly"),
			Sort:      proto.Uint32(3),
			ProxyIp:   proto.String("blhxusproxy.yo-star.com"),
			ProxyPort: proto.Uint32(20001),
		},
		{
			Ids:       []uint32{4},
			Ip:        proto.String("blhxusgs1api.yo-star.com"),
			Port:      proto.Uint32(80),
			State:     proto.Uint32(SERVER_STATE_ONLINE),
			Name:      proto.String("Belfast - @BetterCallMolly"),
			Sort:      proto.Uint32(4),
			ProxyIp:   proto.String("blhxusproxy.yo-star.com"),
			ProxyPort: proto.Uint32(20001),
		},
	}
)

// Answer to a pseudo CS_8239 packet with a SC_8239 packet + server list (HTTP/1.1 200 OK)
func Forge_SC8239(buffer *[]byte, client *connection.Client) (int, int, error) {
	const packetId = 8239
	var answerBuffer bytes.Buffer

	// Write the HTTP header
	answerBuffer.WriteString("HTTP/1.1 200 OK\r\nContent-Type: text/plain;charset=utf-8\r\nAccess-Control-Allow-Origin: \r\nContent-Length: 335\r\n\r\n")

	// Write the JSON-ized server list
	jsonData, err := json.Marshal(Servers)
	if err != nil {
		return 0, packetId, err
	}
	answerBuffer.Write(jsonData)

	// Write buffer to fd
	n, err := syscall.Write(client.FD, answerBuffer.Bytes())
	if err != nil {
		return 0, packetId, err
	}
	return n, packetId, nil
}
