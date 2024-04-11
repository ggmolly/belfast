package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
)

func EquippedWeaponSkin(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_14101
	return client.SendMessage(14101, &response)
}
