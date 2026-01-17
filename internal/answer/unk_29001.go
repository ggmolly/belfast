package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func UNK_29001(buffer *[]byte, client *connection.Client) (int, int, error) {
	// TODO: Bind TB data to the requesting commander.
	var payload protobuf.CS_29001
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 29002, err
	}
	response := protobuf.SC_29002{
		Result:    proto.Uint32(0),
		Tb:        tbInfoPlaceholder(),
		Permanent: tbPermanentPlaceholder(),
	}
	response.Tb.Id = proto.Uint32(payload.GetId())
	return client.SendMessage(29002, &response)
}
