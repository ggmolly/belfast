package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

// ItemOp15004 handles CS_15004 (id, count) -> SC_15005 (result).
//
// Semantics: not observed in available captures (e.g. wg-traffic.pcap) and no
// in-repo send-site exists, so Belfast treats this as a legacy/unused item
// operation and replies with success without mutating state.
func ItemOp15004(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_15004
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 15005, err
	}

	response := protobuf.SC_15005{Result: proto.Uint32(0)}
	return client.SendMessage(15005, &response)
}
