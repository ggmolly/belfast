package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func CompositeSpWeapon(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_14209
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 14209, err
	}

	templateId := data.GetTemplateId()
	response := protobuf.SC_14210{}
	if templateId == 0 || client.Commander == nil {
		response.Result = proto.Uint32(1)
		return client.SendMessage(14210, &response)
	}

	entry := orm.OwnedSpWeapon{
		OwnerID:        client.Commander.CommanderID,
		TemplateID:     templateId,
		Attr1:          0,
		Attr2:          0,
		AttrTemp1:      0,
		AttrTemp2:      0,
		Effect:         0,
		Pt:             0,
		EquippedShipID: 0,
	}
	if err := orm.GormDB.Create(&entry).Error; err != nil {
		return 0, 14210, err
	}
	client.Commander.OwnedSpWeapons = append(client.Commander.OwnedSpWeapons, entry)
	if client.Commander.OwnedSpWeaponsMap != nil {
		// Appending can reallocate the slice and stale existing pointers in the map.
		client.Commander.RebuildOwnedSpWeaponMap()
	}

	response.Result = proto.Uint32(0)
	response.Spweapon = orm.ToProtoOwnedSpWeapon(entry)
	return client.SendMessage(14210, &response)
}
