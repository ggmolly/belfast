package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
)

func CommanderGuildTechnologies(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_62101
	return client.SendMessage(62101, &response)
}
