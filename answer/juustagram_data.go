package answer

import (
	"github.com/ggmolly/belfast/connection"

	"github.com/ggmolly/belfast/protobuf"
)

func JuustagramData(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_11700
	return client.SendMessage(11700, &response)
}
