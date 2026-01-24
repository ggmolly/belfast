package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func NewEducateRequest(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_29001
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29002, err
	}
	state, err := loadEducateState(client, payload.GetId())
	if err != nil {
		return 0, 29002, err
	}
	response := protobuf.SC_29002{
		Result:    proto.Uint32(0),
		Tb:        state.Info,
		Permanent: state.Permanent,
	}
	if err := saveEducateState(state); err != nil {
		return 0, 29002, err
	}
	return client.SendMessage(29002, &response)
}
