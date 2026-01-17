package answer

import (
	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

var validSC60004 protobuf.SC_60004

func GetGuildRequestsCommandResponse(buffer *[]byte, client *connection.Client) (int, int, error) {
	return client.SendMessage(60004, &validSC60004)
}

func init() {
	data := []byte{}
	proto.Unmarshal(data, &validSC60004)
}
