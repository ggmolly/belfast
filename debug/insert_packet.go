package debug

import (
	"fmt"

	"github.com/ggmolly/belfast/logger"
	"github.com/ggmolly/belfast/orm"
)

func InsertPacket(packetId int, payload *[]uint8) {
	if packetId == 8239 {
		return
	}
	dbDebug := orm.Debug{
		PacketSize: len(*payload),
		Data:       *payload,
		PacketID:   packetId,
	}
	err := orm.GormDB.Create(&dbDebug).Error
	if err != nil {
		logger.LogEvent("Debug", "InsertPacket", fmt.Sprintf("Failed to insert packet %d", packetId), logger.LOG_LEVEL_ERROR)
		logger.LogEvent("Debug", "InsertPacket", err.Error(), logger.LOG_LEVEL_ERROR)
	}
}
