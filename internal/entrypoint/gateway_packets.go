package entrypoint

import (
	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/packets"
)

func registerGatewayPackets() {
	packets.RegisterPacketHandler(10800, []packets.PacketHandler{answer.Forge_SC10801})
	packets.RegisterPacketHandler(8239, []packets.PacketHandler{answer.Forge_SC8239})
	packets.RegisterPacketHandler(10018, []packets.PacketHandler{answer.Forge_SC10019})
	packets.RegisterPacketHandler(10020, []packets.PacketHandler{answer.Forge_SC10021_Gateway})
}
