package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func EducateRequest(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_27001{
		Result: proto.Uint32(0),
		Child: &protobuf.CHILD_INFO{
			Tid:        proto.Uint32(1),
			Mood:       proto.Uint32(0),
			Money:      proto.Uint32(0),
			SiteNumber: proto.Uint32(0),
			CurTime: &protobuf.CHILD_TIME{
				Month: proto.Uint32(2),
				Day:   proto.Uint32(7),
				Week:  proto.Uint32(4),
			},
			Favor: &protobuf.CHILD_FAVOR{
				Lv:  proto.Uint32(1),
				Exp: proto.Uint32(0),
			},
			Attrs: []*protobuf.CHILD_ATTR{
				{Id: proto.Uint32(201), Val: proto.Uint32(0)},
				{Id: proto.Uint32(202), Val: proto.Uint32(0)},
				{Id: proto.Uint32(203), Val: proto.Uint32(0)},
			},
			Items:                   []*protobuf.CHILD_ITEM{},
			PlanHistory:             []*protobuf.CHILD_PLAN_HISTORY{},
			Memorys:                 []uint32{},
			Plans:                   []*protobuf.CHILD_PLAN_CELL{},
			Polaroids:               []*protobuf.CHILD_POLAROID{},
			Target:                  proto.Uint32(0),
			Tasks:                   []*protobuf.CHILD_TASK{},
			RealizedWish:            []uint32{},
			Buffs:                   []*protobuf.CHILD_BUFF{},
			UserName:                proto.String("CHILD_USERNAME_SC_27001"),
			SpecEvents:              []uint32{},
			CanTriggerHomeEvent:     proto.Uint32(0),
			HomeEvents:              []uint32{},
			DiscountEventId:         []uint32{},
			Shop:                    []*protobuf.CHILD_SHOP_DATA{},
			OptionRecords:           []*protobuf.CHILD_OPTION_RECORD{},
			FavorAwardHistory:       []uint32{},
			IsEnding:                proto.Uint32(0),
			NewGamePlusCount:        proto.Uint32(0),
			HadTargetStageAward:     proto.Uint32(0),
			HadAdjustment:           proto.Uint32(0),
			IsSpecialSecretaryValid: proto.Uint32(0),
		},
	}
	return client.SendMessage(27001, &response)
}
