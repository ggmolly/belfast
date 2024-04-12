package answer

import (
	"github.com/ggmolly/belfast/connection"

	"github.com/ggmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func PermanentActivites(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_11210
	response.PermanentNow = proto.Uint32(0)
	return client.SendMessage(11210, &response)
}
