package entrypoint

import (
	"testing"

	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/region"
)

func TestPacketRegistryIncludes15010(t *testing.T) {
	region.ResetCurrentForTest()
	packets.PacketDecisionFn = make(map[int][]packets.PacketHandler)
	registerPackets()
	if _, ok := packets.PacketDecisionFn[15010]; !ok {
		t.Fatalf("expected CS_15010 to be registered")
	}
}

func TestPacketRegistryIncludes17109(t *testing.T) {
	region.ResetCurrentForTest()
	packets.PacketDecisionFn = make(map[int][]packets.PacketHandler)
	registerPackets()
	if _, ok := packets.PacketDecisionFn[17109]; !ok {
		t.Fatalf("expected CS_17109 to be registered")
	}
}
