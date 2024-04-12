package answer

import (
	"fmt"

	"github.com/ggmolly/belfast/connection"
	"github.com/ggmolly/belfast/logger"
	"github.com/ggmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func ProposeShip(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_12032
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 12032, err
	}
	logger.LogEvent("Dock", "Propose", fmt.Sprintf("uid=%d has proposed ship id=%d", client.Commander.CommanderID, payload.GetShipId()), logger.LOG_LEVEL_DEBUG)
	return client.SendMessage(12033, client.Commander.ProposeShip(payload.GetShipId()))
}
