package answer

import (
	"log"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func FetchVoteTicketInfo(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_17201
	err := proto.Unmarshal((*buffer), &payload)
	if err != nil {
		return 0, 17202, err
	}

	log.Println("Client asked for ticket info type:", payload.GetType())
	response := protobuf.SC_17202{
		DailyVote:     proto.Uint32(0),
		LoveVote:      proto.Uint32(0),
		DailyShipList: []uint32{},
	}
	return client.SendMessage(17202, &response)
}
