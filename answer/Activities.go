package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
)

func Activities(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_11200
	return client.SendMessage(11200, &response)
}
