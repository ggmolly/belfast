package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func CheaterMark(buffer *[]byte, client *connection.Client) (int, int, error) {
	var protoData protobuf.CS_10994
	err := proto.Unmarshal((*buffer), &protoData)
	if err != nil {
		return 0, 10995, err
	}

	response := protobuf.SC_10995{
		Result: proto.Uint32(protoData.GetType()),
	}

	return client.SendMessage(10995, &response)
}
