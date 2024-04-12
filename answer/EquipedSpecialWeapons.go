package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func EquipedSpecialWeapons(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_14001
	response.SpweaponBagSize = proto.Uint32(0)
	return client.SendMessage(14001, &response)
}
