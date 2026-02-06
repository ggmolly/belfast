package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func GetPhantomQuestProgress(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_12212
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 12213, err
	}

	response := protobuf.SC_12213{}
	shipIDs := payload.GetShipIdList()
	if len(shipIDs) > 0 {
		response.ShipCountList = make([]*protobuf.KVDATA, 0, len(shipIDs))
		for _, shipID := range shipIDs {
			response.ShipCountList = append(response.ShipCountList, &protobuf.KVDATA{
				Key:   proto.Uint32(shipID),
				Value: proto.Uint32(0),
			})
		}
	}

	return client.SendMessage(12213, &response)
}
