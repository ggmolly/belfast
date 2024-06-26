package answer

import (
	"log"

	"github.com/ggmolly/belfast/connection"

	"github.com/ggmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func UNK_17204(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_17203
	err := proto.Unmarshal((*buffer), &payload)
	if err != nil {
		return 0, 17204, err
	}
	var response protobuf.SC_17204
	log.Println("Client asked for type:", payload.GetType())
	return client.SendMessage(17204, &response)
}
