package packets

import (
	"fmt"
	"log"
	"os"

	"github.com/ggmolly/belfast/connection"
	"github.com/ggmolly/belfast/debug"
	"github.com/ggmolly/belfast/logger"
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
	switch os.Getenv("AL_REGION") {
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
		log.Fatalf("could not find region %s to register localized packet handler", os.Getenv("AL_REGION"))
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
