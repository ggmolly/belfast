package answer

import (
	"github.com/ggmolly/belfast/connection"

	"github.com/ggmolly/belfast/protobuf"
)

func CommanderGuildTechnologies(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_62101
	return client.SendMessage(62101, &response)
}
