package answer

import (
	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func EventData(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_26120
	response.WeeklyFree = proto.Uint32(0)
	response.MonthlyTicket = proto.Uint32(0)
	response.PayCoinCount = proto.Uint32(0)
	response.FirstEnter = proto.Uint32(0)
	return client.SendMessage(26120, &response)
}
