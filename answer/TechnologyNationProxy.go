package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
)

func TechnologyNationProxy(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_64000
	return client.SendMessage(64000, &response)
}
