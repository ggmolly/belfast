package answer

import (
	"github.com/bettercallmolly/belfast/connection"
	"github.com/bettercallmolly/belfast/protobuf"
)

func CommanderFriendList(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_50000 // Create an empty FriendList / RequestList
	return client.SendMessage(50000, &response)
}
