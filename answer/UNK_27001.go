package answer

import (
	"github.com/bettercallmolly/belfast/connection"
	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func UNK_27001(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_27001{
		Result: proto.Uint32(0),
		Child: &protobuf.CHILD_INFO{
			Tid:        proto.Uint32(0),
			Mood:       proto.Uint32(0),
			Money:      proto.Uint32(0),
			SiteNumber: proto.Uint32(0),
			CurTime: &protobuf.CHILD_TIME{
				Month: proto.Uint32(0),
				Day:   proto.Uint32(0),
				Week:  proto.Uint32(0),
			},
			Favor: &protobuf.CHILD_FAVOR{
				Lv:  proto.Uint32(0),
				Exp: proto.Uint32(0),
			},
			Target:                  proto.Uint32(0),
			UserName:                proto.String("CHILD_USERNAME_SC_27001"),
			CanTriggerHomeEvent:     proto.Uint32(0),
			IsEnding:                proto.Uint32(0),
			NewGamePlusCount:        proto.Uint32(0),
			HadTargetStageAward:     proto.Uint32(0),
			HadAdjustment:           proto.Uint32(0),
			IsSpecialSecretaryValid: proto.Uint32(0),
		},
	}
	return client.SendMessage(27001, &response)
}
