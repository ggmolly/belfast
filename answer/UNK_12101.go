package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

var validSC12101 protobuf.SC_12101

func UNK_12101(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_12101
	response.GroupList = append(response.GroupList, &protobuf.GROUPINFO{
		Id:       proto.Uint32(1),
		ShipList: []uint32{307081, 405051, 901111, 605021, 805011, 603021},
	})
	return client.SendMessage(12101, &response)
}

func init() {
	data := []byte{}
	panic("replayed packet: replace this with the actual data")
	proto.Unmarshal(data, &validSC12101)
}
