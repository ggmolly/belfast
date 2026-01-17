package answer

import (
	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func GuildGetUserInfoCommand(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_60103{
		UserInfo: &protobuf.USER_GUILD_INFO{
			DonateCount:    proto.Uint32(0),
			BenefitTime:    proto.Uint32(0),
			WeeklyTaskFlag: proto.Uint32(0),
			ExtraDonate:    proto.Uint32(0),
			ExtraOperation: proto.Uint32(0),
		},
	}
	return client.SendMessage(60103, &response)
}
