package entrypoint

import (
	"testing"

	"github.com/ggmolly/belfast/internal/packets"
)

func TestRegisterPacketsIncludes14004(t *testing.T) {
	packets.PacketDecisionFn = make(map[int][]packets.PacketHandler)
	registerPackets()
	if _, ok := packets.PacketDecisionFn[14004]; !ok {
		t.Fatalf("expected handler for CS_14004 to be registered")
	}
}
