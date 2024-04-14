package answer

import (
	"github.com/ggmolly/belfast/connection"
	"github.com/ggmolly/belfast/protobuf"
)

func CommanderFriendList(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_50000 // Create an empty FriendList / RequestList
	return client.SendMessage(50000, &response)
}
