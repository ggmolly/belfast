package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func CompensateNotification(buffer *[]byte, client *connection.Client) (int, int, error) {
	// TODO: Populate compensation reward metadata from storage.
	response := protobuf.SC_30101{
		Number:       proto.Uint32(0),
		MaxTimestamp: proto.Uint32(0),
	}
	return client.SendMessage(30101, &response)
}
