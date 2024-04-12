package packets

import (
	"fmt"

	"github.com/ggmolly/belfast/connection"
	"github.com/ggmolly/belfast/debug"
	"github.com/ggmolly/belfast/logger"
)

const (
	HEADER_SIZE = 7
)

type PacketHandler func(*[]byte, *connection.Client) (int, int, error)

var PacketDecisionFn = map[int][]PacketHandler{}

func RegisterPacketHandler(packetId int, handlers []PacketHandler) {
	logger.LogEvent("Handler", "Added", fmt.Sprintf("CS_%d", packetId), logger.LOG_LEVEL_DEBUG)
	PacketDecisionFn[packetId] = handlers
}

// Find each packet in the buffer and dispatch it to the appropriate handler.
func Dispatch(buffer *[]byte, client *connection.Client) {
	offset := 0
	for offset < len(*buffer) {
		packetId := GetPacketId(offset, buffer)
		packetSize := GetPacketSize(offset, buffer) + 2
		client.PacketIndex = GetPacketIndex(offset, buffer)
		handlers, ok := PacketDecisionFn[packetId]
		headerlessBuffer := (*buffer)[offset+HEADER_SIZE:]
		if !ok {
			logger.LogEvent("Handler", "Missing", fmt.Sprintf("CS_%d", packetId), logger.LOG_LEVEL_ERROR)
			debug.InsertPacket(packetId, &headerlessBuffer)
		} else {
			debug.InsertPacket(packetId, &headerlessBuffer)
			for _, handler := range handlers {
				// offset buffer by header size
				_, packetId, err := handler(&headerlessBuffer, client)
				if err != nil {
					logger.LogEvent("Handler", "Error", fmt.Sprintf("SC_%d - %v", packetId, err), logger.LOG_LEVEL_ERROR)
				}
			}
		}
		offset += packetSize
	}
	client.Flush()
}
