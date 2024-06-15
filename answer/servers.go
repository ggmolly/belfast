package answer

import (
	"bytes"
	"encoding/json"

	"github.com/ggmolly/belfast/connection"
	"github.com/ggmolly/belfast/protobuf"
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

	n, err := (*client.Connection).Write(answerBuffer.Bytes())
	if err != nil {
		return 0, packetId, err
	}
	return n, packetId, nil
}
