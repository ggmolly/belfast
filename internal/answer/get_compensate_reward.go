package answer

import (
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func GetCompensateReward(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_30104
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 30104, err
	}
	response := protobuf.SC_30105{Result: proto.Uint32(1)}
	compensation := client.Commander.CompensationsMap[payload.GetRewardId()]
	if compensation == nil {
		return client.SendMessage(30105, &response)
	}
	now := time.Now()
	if compensation.IsExpired(now) || compensation.AttachFlag {
		return client.SendMessage(30105, &response)
	}
	attachments, err := compensation.CollectAttachments(client.Commander)
	if err != nil {
		return 0, 30104, err
	}
	response.DropList = orm.ToProtoCompensationDropInfoList(attachments)
	count, maxTimestamp := orm.CompensationSummary(client.Commander.Compensations, now)
	response.Number = proto.Uint32(count)
	response.MaxTimestamp = proto.Uint32(maxTimestamp)
	response.Result = proto.Uint32(0)
	return client.SendMessage(30105, &response)
}
