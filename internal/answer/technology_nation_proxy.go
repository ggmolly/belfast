package answer

import (
	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
)

func TechnologyNationProxy(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_64000
	return client.SendMessage(64000, &response)
}
