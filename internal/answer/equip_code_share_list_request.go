package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func EquipCodeShareListRequest(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_17601
	err := proto.Unmarshal(*buffer, &data)
	if err != nil {
		return 0, 17601, err
	}

	// Phase 1: unblock the UI by replying successfully with empty lists.
	_ = data.GetShipgroup()
	response := protobuf.SC_17602{
		Result:      proto.Uint32(0),
		Infos:       []*protobuf.EQCODE_SHARE_INFO{},
		RecentInfos: []*protobuf.EQCODE_SHARE_INFO{},
	}

	return client.SendMessage(17602, &response)
}
