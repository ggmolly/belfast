package answer

import (
	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
)

func EquippedWeaponSkin(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_14101
	return client.SendMessage(14101, &response)
}
