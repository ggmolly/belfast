package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func GameTracking(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_10991
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 10992, err
	}
	// TODO: Persist or forward tracking payloads once analytics storage exists.
	response := protobuf.CS_10992{
		TrackType: proto.Uint32(0),
		EventId:   proto.Uint32(0),
		Para1:     proto.String(""),
		Para2:     proto.String(""),
		Para3:     proto.String(""),
	}
	return client.SendMessage(10992, &response)
}
