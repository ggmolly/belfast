package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ShipAction12029(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_12029
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 12030, err
	}

	response := protobuf.SC_12030{
		Result:   proto.Uint32(0),
		ShipList: []*protobuf.SHIPINFO{},
	}

	return client.SendMessage(12030, &response)
}
