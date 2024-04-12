package answer

import (
	"github.com/ggmolly/belfast/connection"

	"github.com/ggmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

// XXX: i have no idea what this packet is for but this works, so...
func BuildFinish(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_12043
	err := proto.Unmarshal(*buffer, &data)
	if err != nil {
		return 0, 12043, err
	}

	response := protobuf.SC_12044{
		InfoList: []*protobuf.BUILD_INFO{ // ???
			&protobuf.BUILD_INFO{
				Pos: proto.Uint32(1),
				Tid: proto.Uint32(1),
			},
		},
	}
	return client.SendMessage(12044, &response)
}
