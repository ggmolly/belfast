package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ConfirmReforgeSpWeapon(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_14207
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 14208, err
	}
	response := protobuf.SC_14208{Result: proto.Uint32(1)}

	if client.Commander == nil || client.Commander.OwnedSpWeaponsMap == nil {
		return client.SendMessage(14208, &response)
	}
	spweaponID := data.GetSpweaponId()
	spweapon, ok := client.Commander.OwnedSpWeaponsMap[spweaponID]
	if !ok {
		return client.SendMessage(14208, &response)
	}

	shipID := data.GetShipId()
	if shipID != 0 {
		if _, ok := client.Commander.OwnedShipsMap[shipID]; !ok {
			return client.SendMessage(14208, &response)
		}
		if spweapon.EquippedShipID != 0 && spweapon.EquippedShipID != shipID {
			return client.SendMessage(14208, &response)
		}
	}

	switch data.GetCmd() {
	case 0:
		// discard: keep base attrs unchanged
	case 1:
		// exchange: apply temp attrs to base
		spweapon.Attr1 = spweapon.AttrTemp1
		spweapon.Attr2 = spweapon.AttrTemp2
	default:
		return client.SendMessage(14208, &response)
	}

	spweapon.AttrTemp1 = 0
	spweapon.AttrTemp2 = 0
	if err := orm.GormDB.Save(spweapon).Error; err != nil {
		return 0, 14208, err
	}
	response.Result = proto.Uint32(0)
	return client.SendMessage(14208, &response)
}
