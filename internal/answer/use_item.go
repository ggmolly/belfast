package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func UseItem(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_15002
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 15003, err
	}
	if client.Commander.CommanderItemsMap == nil && client.Commander.MiscItemsMap == nil {
		logger.LogEvent("Item", "Use", "commander maps missing, reloading commander", logger.LOG_LEVEL_INFO)
		if err := client.Commander.Load(); err != nil {
			logger.LogEvent("Item", "Use", "commander load failed", logger.LOG_LEVEL_ERROR)
			return 0, 15003, err
		}
	}
	response := protobuf.SC_15003{Result: proto.Uint32(1)}
	outcome, err := useItem(client, &payload)
	if err != nil {
		return 0, 15003, err
	}
	if outcome != nil {
		response.Result = proto.Uint32(outcome.result)
		response.DropList = outcome.dropList
	}
	return client.SendMessage(15003, &response)
}
