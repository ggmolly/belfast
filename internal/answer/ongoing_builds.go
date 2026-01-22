package answer

import (
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/orm"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func OngoingBuilds(buffer *[]byte, client *connection.Client) (int, int, error) {
	now := time.Now()
	orderedBuilds := orm.OrderedBuilds(client.Commander.Builds)
	buildInfos := make([]*protobuf.BUILDINFO, len(orderedBuilds))
	for i, work := range orderedBuilds {
		buildInfos[i] = orm.ToProtoBuildInfo(orm.BuildInfoPayload{
			Build:      &work,
			PoolID:     work.PoolID,
			BuildTime:  orm.RemainingSeconds(work.FinishesAt, now),
			FinishTime: work.FinishesAt,
		})
	}
	response := protobuf.SC_12024{
		WorklistCount: proto.Uint32(consts.MaxBuildWorkCount),
		WorklistList:  buildInfos,
		DrawCount_1:   proto.Uint32(client.Commander.DrawCount1),
		DrawCount_10:  proto.Uint32(client.Commander.DrawCount10),
		ExchangeCount: proto.Uint32(client.Commander.ExchangeCount),
	}
	return client.SendMessage(12024, &response)
}
