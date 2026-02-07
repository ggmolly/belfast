package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func EquipedSpecialWeapons(buffer *[]byte, client *connection.Client) (int, int, error) {
	entries, err := orm.ListConfigEntries(orm.GormDB, "ShareCfg/spweapon_data_statistics.json")
	if err != nil {
		return 0, 14001, err
	}
	response := protobuf.SC_14001{
		SpweaponBagSize: proto.Uint32(uint32(len(entries))),
	}
	response.SpweaponList = orm.ToProtoOwnedSpWeaponList(client.Commander.OwnedSpWeapons)
	response.EquipList = make([]*protobuf.EQUIPINFO, 0, len(client.Commander.OwnedEquipments))
	for _, owned := range client.Commander.OwnedEquipments {
		if owned.Count == 0 {
			continue
		}
		response.EquipList = append(response.EquipList, &protobuf.EQUIPINFO{
			Id:    proto.Uint32(owned.EquipmentID),
			Count: proto.Uint32(owned.Count),
		})
	}
	return client.SendMessage(14001, &response)
}
