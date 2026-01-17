package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"google.golang.org/protobuf/proto"

	"github.com/ggmolly/belfast/internal/protobuf"
)

func CommanderGuildData(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_60000{
		Guild: &protobuf.GUILD_INFO{
			Base: &protobuf.GUILD_BASE_INFO{
				Id:              proto.Uint32(0),
				Policy:          proto.Uint32(0),
				Faction:         proto.Uint32(0),
				Name:            proto.String(""),
				Level:           proto.Uint32(0),
				Announce:        proto.String(""),
				Manifesto:       proto.String(""),
				Exp:             proto.Uint32(0),
				MemberCount:     proto.Uint32(0),
				ChangeFactionCd: proto.Uint32(0),
				KickLeaderCd:    proto.Uint32(0),
			},
			GuildEx: &protobuf.GUILD_EXPANSION_INFO{
				Capital: proto.Uint32(0),
				ThisWeeklyTasks: &protobuf.WEEKLY_TASK{
					Id:            proto.Uint32(0),
					Progress:      proto.Uint32(0),
					Monday_0Clock: proto.Uint32(0),
				},
				BenefitFinishTime:     proto.Uint32(0),
				RetreatCnt:            proto.Uint32(0),
				TechCancelCnt:         proto.Uint32(0),
				LastBenefitFinishTime: proto.Uint32(0),
				ActiveEventCnt:        proto.Uint32(0),
			},
		},
	}
	return client.SendMessage(60000, &response)
}
