package answer

import (
	"log"

	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func FetchVoteInfo(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_17203
	err := proto.Unmarshal((*buffer), &payload)
	if err != nil {
		return 0, 17204, err
	}
	var response protobuf.SC_17204
	log.Println("Client asked for type:", payload.GetType())
	return client.SendMessage(17204, &response)
}
