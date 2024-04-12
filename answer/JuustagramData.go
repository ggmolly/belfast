package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
)

func JuustagramData(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_11700
	return client.SendMessage(11700, &response)
}
