package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func FinishPhantomQuest(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_12210
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 12211, err
	}
	response := protobuf.SC_12211{Result: proto.Uint32(0)}
	shipID := payload.GetShipId()
	shadowID := payload.GetSkinShadowId()
	ship, ok := client.Commander.OwnedShipsMap[shipID]
	if !ok {
		response.Result = proto.Uint32(1)
		return client.SendMessage(12211, &response)
	}
	quest, err := orm.GetTechnologyShadowUnlockConfig(shadowID)
	if err != nil {
		response.Result = proto.Uint32(1)
		return client.SendMessage(12211, &response)
	}
	skinID, err := orm.GetShipBaseSkinIDTx(nil, ship.ShipID)
	if err != nil || skinID == 0 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(12211, &response)
	}

	ownedSkins, err := orm.ListOwnedShipShadowSkins(client.Commander.CommanderID, []uint32{shipID})
	if err != nil {
		response.Result = proto.Uint32(1)
		return client.SendMessage(12211, &response)
	}
	for _, existing := range ownedSkins[shipID] {
		if existing.ShadowID != shadowID {
			continue
		}
		if existing.SkinID != skinID {
			if err := orm.UpsertOwnedShipShadowSkin(nil, client.Commander.CommanderID, shipID, shadowID, skinID); err != nil {
				response.Result = proto.Uint32(1)
				return client.SendMessage(12211, &response)
			}
		}
		return client.SendMessage(12211, &response)
	}

	if quest.Type == 5 {
		if err := client.Commander.ConsumeResource(4, quest.TargetNum); err != nil {
			response.Result = proto.Uint32(1)
			return client.SendMessage(12211, &response)
		}
	}
	if err := orm.UpsertOwnedShipShadowSkin(nil, client.Commander.CommanderID, shipID, shadowID, skinID); err != nil {
		response.Result = proto.Uint32(1)
		return client.SendMessage(12211, &response)
	}
	return client.SendMessage(12211, &response)
}
