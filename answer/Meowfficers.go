package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func Meowfficers(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_25001
	response.UsageCount = proto.Uint32(0)
	return client.SendMessage(25001, &response)
}
