package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func FinishPhantomQuest(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_12210
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 12211, err
	}
	response := protobuf.SC_12211{Result: proto.Uint32(0)}

	shipID := data.GetShipId()
	shadowID := data.GetSkinShadowId()
	if _, ok := client.Commander.OwnedShipsMap[shipID]; ok {
		entry := orm.OwnedShipSkinShadow{
			CommanderID: client.Commander.CommanderID,
			ShipID:      shipID,
			ShadowID:    shadowID,
			SkinID:      0,
		}
		if err := orm.UpsertOwnedShipSkinShadow(orm.GormDB, &entry); err != nil {
			response.Result = proto.Uint32(1)
		}
	}

	return client.SendMessage(12211, &response)
}
