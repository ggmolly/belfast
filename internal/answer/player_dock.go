package answer

import (
	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

var validSC12001 protobuf.SC_12001

func boolToUint32(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}

func PlayerDock(buffer *[]byte, client *connection.Client) (int, int, error) {
	// Send first 100 ships
	maxSlice := 101
	if len(client.Commander.Ships) < maxSlice {
		maxSlice = len(client.Commander.Ships)
	}
	var shipList []*protobuf.SHIPINFO
	for _, ship := range client.Commander.Ships[:maxSlice] {
		shipList = append(shipList, &protobuf.SHIPINFO{
			Id:                  proto.Uint32(ship.ID),
			TemplateId:          proto.Uint32(ship.ShipID),
			Level:               proto.Uint32(ship.Level),
			Exp:                 proto.Uint32(0),
			Energy:              proto.Uint32(ship.Energy),
			IsLocked:            proto.Uint32(boolToUint32(ship.IsLocked)),
			Intimacy:            proto.Uint32(ship.Intimacy),
			Proficiency:         proto.Uint32(boolToUint32(ship.Proficiency)),
			SkinId:              proto.Uint32(ship.SkinID),
			Propose:             proto.Uint32(boolToUint32(ship.Propose)),
			Name:                proto.String(ship.CustomName),
			MaxLevel:            proto.Uint32(ship.MaxLevel),
			BluePrintFlag:       proto.Uint32(boolToUint32(ship.BlueprintFlag)),
			CommonFlag:          proto.Uint32(boolToUint32(ship.CommonFlag)),
			ActivityNpc:         proto.Uint32(ship.ActivityNPC),
			CreateTime:          proto.Uint32(uint32(ship.CreateTime.Unix())),
			ChangeNameTimestamp: proto.Uint32(0),
			State: &protobuf.SHIPSTATE{
				State: proto.Uint32(1),
			},
			Commanderid:    proto.Uint32(0),
			EquipInfoList:  nil,
			TransformList:  nil,
			SkillIdList:    nil,
			StrengthList:   nil,
			MetaRepairList: nil,
			CoreList:       nil,
			Spweapon:       nil,
		})
	}
	validSC12001.Shiplist = shipList
	return client.SendMessage(12001, &validSC12001)
}
