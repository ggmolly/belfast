package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func UNK_34502(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_34502{
		FightCount:           proto.Uint32(0),
		FightCountUpdateTime: proto.Uint32(0),
		SelfBoss: &protobuf.WORLDBOSS_INFO{
			Id:         proto.Uint32(0),
			TemplateId: proto.Uint32(0),
			Lv:         proto.Uint32(0),
			Hp:         proto.Uint32(0),
			Owner:      proto.Uint32(0),
			LastTime:   proto.Uint32(0),
			KillTime:   proto.Uint32(0),
			FightCount: proto.Uint32(0),
			RankCount:  proto.Uint32(0),
		},
		SummonPt:            proto.Uint32(0),
		SummonPtOld:         proto.Uint32(0),
		SummonPtDailyAcc:    proto.Uint32(0),
		SummonPtOldDailyAcc: proto.Uint32(0),
		SummonFree:          proto.Uint32(0),
		AutoFightFinishTime: proto.Uint32(0),
		DefaultBossId:       proto.Uint32(0),
		AutoFightMaxDamage:  proto.Uint32(0),
		GuildSupport:        proto.Uint32(0),
		FriendSupport:       proto.Uint32(0),
		WorldSupport:        proto.Uint32(0),
	}
	return client.SendMessage(34502, &response)
}
