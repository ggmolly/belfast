package answer

import (
	"github.com/ggmolly/belfast/connection"

	"github.com/ggmolly/belfast/protobuf"
)

func EquippedWeaponSkin(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_14101
	return client.SendMessage(14101, &response)
}
