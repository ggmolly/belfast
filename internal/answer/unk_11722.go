package answer

import (
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func UNK_11722(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_11723{
		ResultList: []uint32{},
		OpTime:     proto.Uint32(uint32(time.Now().Unix())),
	}
	return client.SendMessage(11723, &response)
}
