package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

// Get the corresponding template_id in the ship build list
func BuildFinish(buffer *[]byte, client *connection.Client) (int, int, error) {
	orderedBuilds := orm.OrderedBuilds(client.Commander.Builds)
	buildInfos := make([]*protobuf.BUILD_INFO, len(orderedBuilds))
	for i, work := range orderedBuilds {
		buildInfos[i] = &protobuf.BUILD_INFO{
			Pos: proto.Uint32(uint32(i + 1)),
			Tid: proto.Uint32(work.ShipID),
		}
	}

	response := protobuf.SC_12044{
		InfoList: buildInfos,
	}

	return client.SendMessage(12044, &response)
}
