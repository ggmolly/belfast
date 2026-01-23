package answer

import (
	"encoding/json"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

type equipSkinTemplate struct {
	ID uint32 `json:"id"`
}

func EquippedWeaponSkin(buffer *[]byte, client *connection.Client) (int, int, error) {
	entries, err := orm.ListConfigEntries(orm.GormDB, "ShareCfg/equip_skin_template.json")
	if err != nil {
		return 0, 14101, err
	}
	response := protobuf.SC_14101{
		EquipSkinList: make([]*protobuf.EQUIPSKININFO, 0, len(entries)),
	}
	for _, entry := range entries {
		var template equipSkinTemplate
		if err := json.Unmarshal(entry.Data, &template); err != nil {
			return 0, 14101, err
		}
		response.EquipSkinList = append(response.EquipSkinList, &protobuf.EQUIPSKININFO{
			Id:    proto.Uint32(template.ID),
			Count: proto.Uint32(0),
		})
	}
	return client.SendMessage(14101, &response)
}
