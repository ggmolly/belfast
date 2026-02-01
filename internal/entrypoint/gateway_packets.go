package entrypoint

import (
	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/packets"
)

func registerGatewayPackets() {
	packets.RegisterPacketHandler(10800, []packets.PacketHandler{answer.Forge_SC10801_Gateway})
	packets.RegisterPacketHandler(10700, []packets.PacketHandler{answer.GatewayPackInfo})
	packets.RegisterPacketHandler(8239, []packets.PacketHandler{answer.Forge_SC8239})
	packets.RegisterPacketHandler(10018, []packets.PacketHandler{answer.Forge_SC10019})
	packets.RegisterPacketHandler(10001, []packets.PacketHandler{answer.RegisterAccount})
	packets.RegisterPacketHandler(10020, []packets.PacketHandler{answer.Forge_SC10021_Gateway})
}
