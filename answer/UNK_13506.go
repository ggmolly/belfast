package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
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
