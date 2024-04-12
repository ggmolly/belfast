package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
)

func CommanderGuildChat(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_60101
	return client.SendMessage(60101, &response)
}
