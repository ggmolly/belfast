package answer

import (
	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

var validSC13506 protobuf.SC_13506

func UNK_13506(buffer *[]byte, client *connection.Client) (int, int, error) {
	return client.SendMessage(13506, &validSC13506)
}

func init() {
	data := []byte{}
	proto.Unmarshal(data, &validSC13506)
}
