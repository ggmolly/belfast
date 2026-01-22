package answer

import (
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func compensationToTimeRewardInfo(compensation *orm.Compensation) *protobuf.TIME_REWARD_INFO {
	return &protobuf.TIME_REWARD_INFO{
		Id:             proto.Uint32(compensation.ID),
		Timestamp:      proto.Uint32(uint32(compensation.ExpiresAt.Unix())),
		Title:          proto.String(compensation.Title),
		Text:           proto.String(compensation.Text),
		AttachmentList: orm.ToProtoCompensationDropInfoList(compensation.Attachments),
		AttachFlag:     proto.Uint32(boolToUint32(compensation.AttachFlag)),
		SendTime:       proto.Uint32(uint32(compensation.SendTime.Unix())),
	}
}

func GetCompensateList(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_30102
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 30102, err
	}
	response := protobuf.SC_30103{}
	now := time.Now()
	for i := range client.Commander.Compensations {
		compensation := &client.Commander.Compensations[i]
		if compensation.IsExpired(now) {
			continue
		}
		response.TimeRewardList = append(response.TimeRewardList, compensationToTimeRewardInfo(compensation))
	}
	return client.SendMessage(30103, &response)
}
