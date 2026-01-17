package answer

import (
	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func UNK_13002(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_13002
	response.MaxTeam = proto.Uint32(0)
	return client.SendMessage(13002, &response)
}
