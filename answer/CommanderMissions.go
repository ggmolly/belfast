package answer

import (
	"github.com/bettercallmolly/belfast/connection"
	"github.com/bettercallmolly/belfast/protobuf"
)

func CommanderMissions(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_20001
	return client.SendMessage(20001, &response)
}
