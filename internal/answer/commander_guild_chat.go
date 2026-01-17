package answer

import (
	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
)

func CommanderGuildChat(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_60101
	return client.SendMessage(60101, &response)
}
