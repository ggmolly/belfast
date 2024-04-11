package answer

import (
	"time"

	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

var validSC12024 protobuf.SC_12024

func OngoingBuilds(buffer *[]byte, client *connection.Client) (int, int, error) {
	buildInfos := make([]*protobuf.BUILDINFO, len(client.Commander.Builds))
	for i, work := range client.Commander.Builds {
		buildInfos[i] = &protobuf.BUILDINFO{
			// Time is the number of seconds between now and the finish time
			Time:       proto.Uint32(uint32(time.Until(work.FinishesAt).Seconds())),
			FinishTime: proto.Uint32(uint32(work.FinishesAt.Unix())),
			BuildId:    proto.Uint32(uint32(i + 1)),
		}
	}
	response := protobuf.SC_12024{
		WorklistCount: proto.Uint32(13371337), // TODO: return the number of works depending on user's lvl
		WorklistList:  buildInfos,
		DrawCount_1:   proto.Uint32(2000), // NOTE: these seems to be unused
		DrawCount_10:  proto.Uint32(2000),
		ExchangeCount: proto.Uint32(client.Commander.ExchangeCount),
	}
	return client.SendMessage(12024, &response)
}

func init() {
	data := []byte{0x08, 0x02, 0x18, 0x00, 0x20, 0x00}
	proto.Unmarshal(data, &validSC12024)
}
