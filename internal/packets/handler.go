package packets

import (
	"fmt"
	"log"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/debug"
	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/misc"
)

const (
	HEADER_SIZE = 7
)

type PacketHandler func(*[]byte, *connection.Client) (int, int, error)

var PacketDecisionFn = map[int][]PacketHandler{}

type LocalizedHandler struct {
	CN      *[]PacketHandler
	EN      *[]PacketHandler
	JP      *[]PacketHandler
	KR      *[]PacketHandler
	TW      *[]PacketHandler
	Default *[]PacketHandler
}

// Registers a region agnostic packet handler.
func RegisterPacketHandler(packetId int, handlers []PacketHandler) {
	logger.LogEvent("Handler", "Added", fmt.Sprintf("CS_%d", packetId), logger.LOG_LEVEL_DEBUG)
	PacketDecisionFn[packetId] = handlers
}

// Registers a localized packet handler, will call specific handler(s) based on
// the server's region.
func RegisterLocalizedPacketHandler(packetId int, localizedHandler LocalizedHandler) {
	switch misc.GetSpecifiedRegion() {
	case "CN":
		if localizedHandler.CN != nil {
			PacketDecisionFn[packetId] = *localizedHandler.CN
		}
	case "EN":
		if localizedHandler.EN != nil {
			PacketDecisionFn[packetId] = *localizedHandler.EN
		}
	case "JP":
		if localizedHandler.JP != nil {
			PacketDecisionFn[packetId] = *localizedHandler.JP
		}
	case "KR":
		if localizedHandler.KR != nil {
			PacketDecisionFn[packetId] = *localizedHandler.KR
		}
	case "TW":
		if localizedHandler.TW != nil {
			PacketDecisionFn[packetId] = *localizedHandler.TW
		}
	default:
		log.Fatalf("could not find region %s to register localized packet handler", misc.GetSpecifiedRegion())
	}
}

// Find each packet in the buffer and dispatch it to the appropriate handler.
func Dispatch(buffer *[]byte, client *connection.Client, n int) {
	offset := 0
	for offset < n {
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
				start := time.Now()
				_, packetId, err := handler(&headerlessBuffer, client)
				elapsed := time.Since(start)
				logger.LogEvent("Metrics", "HandlerMs", fmt.Sprintf("CS_%d -> %s", packetId, elapsed), logger.LOG_LEVEL_DEBUG)
				if err != nil {
					client.RecordHandlerError()
					logger.LogEvent("Handler", "Error", fmt.Sprintf("SC_%d - %v", packetId, err), logger.LOG_LEVEL_ERROR)
					client.CloseWithError(err)
					return
				}
			}
		}
		offset += packetSize
	}
	if err := client.Flush(); err != nil {
		logger.LogEvent("Handler", "Flush", fmt.Sprintf("%s:%d -> %v", client.IP, client.Port, err), logger.LOG_LEVEL_ERROR)
	}
}
