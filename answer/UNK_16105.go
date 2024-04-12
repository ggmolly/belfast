package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
)

func UNK_16105(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_16105
	return client.SendMessage(16105, &response)
}
