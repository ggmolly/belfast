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
	return client.SendMessage(14001, &response)
}
