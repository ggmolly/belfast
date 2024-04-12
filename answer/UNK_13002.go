package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func UNK_13002(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_13002
	response.MaxTeam = proto.Uint32(0)
	return client.SendMessage(13002, &response)
}
