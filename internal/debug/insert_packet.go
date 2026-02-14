package debug

import (
	"fmt"

	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/orm"
)

func InsertPacket(packetId int, payload *[]uint8) {
	if packetId == 8239 {
		return
	}
	err := orm.InsertDebugPacket(len(*payload), packetId, *payload)
	if err != nil {
		logger.LogEvent("Debug", "InsertPacket", fmt.Sprintf("Failed to insert packet %d", packetId), logger.LOG_LEVEL_ERROR)
		logger.LogEvent("Debug", "InsertPacket", err.Error(), logger.LOG_LEVEL_ERROR)
	}
}
