package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func QuickExchangeBlueprint(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_15012
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 15013, err
	}
	response := protobuf.SC_15013{}
	if len(payload.UseList) > 0 {
		response.RetList = make([]*protobuf.SC_15003, 0, len(payload.UseList))
	}
	for _, entry := range payload.UseList {
		ret := &protobuf.SC_15003{Result: proto.Uint32(1)}
		outcome, err := useItem(client, entry)
		if err != nil {
			return 0, 15013, err
		}
		if outcome != nil {
			ret.Result = proto.Uint32(outcome.result)
			ret.DropList = outcome.dropList
		}
		response.RetList = append(response.RetList, ret)
	}
	return client.SendMessage(15013, &response)
}
