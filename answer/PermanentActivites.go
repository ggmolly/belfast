package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func PermanentActivites(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_11210
	response.PermanentNow = proto.Uint32(0)
	return client.SendMessage(11210, &response)
}
