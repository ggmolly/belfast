package answer

import (
	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
)

func Activities(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_11200
	return client.SendMessage(11200, &response)
}
