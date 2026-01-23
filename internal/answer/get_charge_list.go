package answer

import (
	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
)

func GetChargeList(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_16105
	return client.SendMessage(16105, &response)
}
