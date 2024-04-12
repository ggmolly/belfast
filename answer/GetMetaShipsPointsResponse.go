package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
)

func GetMetaShipsPointsResponse(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_34002
	return client.SendMessage(34002, &response)
}
