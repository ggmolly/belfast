package answer

import (
	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
)

func CommanderGuildTechnologies(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_62101
	return client.SendMessage(62101, &response)
}
