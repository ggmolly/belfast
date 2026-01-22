package answer

import (
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func CompensateNotification(buffer *[]byte, client *connection.Client) (int, int, error) {
	count, maxTimestamp := orm.CompensationSummary(client.Commander.Compensations, time.Now())
	response := protobuf.SC_30101{
		Number:       proto.Uint32(count),
		MaxTimestamp: proto.Uint32(maxTimestamp),
	}
	return client.SendMessage(30101, &response)
}
