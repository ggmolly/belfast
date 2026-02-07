package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ReportShipEvaluation(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_17109
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 17110, err
	}

	response := protobuf.SC_17110{Result: proto.Uint32(0)}
	return client.SendMessage(17110, &response)
}
