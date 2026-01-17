package answer

import (
	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func EquippedWeaponSkin(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_14101{
		EquipSkinList: []*protobuf.EQUIPSKININFO{
			{
				Id:    proto.Uint32(0),
				Count: proto.Uint32(0),
			},
		},
	}
	return client.SendMessage(14101, &response)
}
