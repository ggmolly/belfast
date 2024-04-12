package answer

import (
	"github.com/ggmolly/belfast/connection"

	"github.com/ggmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

// ???
func UNK_12101(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_12101
	response.GroupList = append(response.GroupList, &protobuf.GROUPINFO{
		Id:       proto.Uint32(1),
		ShipList: []uint32{307081, 405051, 901111, 605021, 805011, 603021},
	})
	return client.SendMessage(12101, &response)
}
