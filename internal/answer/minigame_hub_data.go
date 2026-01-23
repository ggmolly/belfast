package answer

import (
	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

var validSC26102 protobuf.SC_26102

func MiniGameHubData(buffer *[]byte, client *connection.Client) (int, int, error) {
	return client.SendMessage(26102, &validSC26102)
}

func init() {
	data := []byte{}
	proto.Unmarshal(data, &validSC26102)
}
