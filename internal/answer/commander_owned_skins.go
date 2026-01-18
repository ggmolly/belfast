package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func CommanderOwnedSkins(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_12201
	response.SkinList = make([]*protobuf.IDTIMEINFO, len(client.Commander.OwnedSkins))
	for i, skin := range client.Commander.OwnedSkins {
		var expiryTimestamp uint32
		if skin.ExpiresAt != nil {
			expiryTimestamp = uint32(skin.ExpiresAt.Unix())
		}
		response.SkinList[i] = &protobuf.IDTIMEINFO{
			Id:   proto.Uint32(skin.SkinID),
			Time: proto.Uint32(expiryTimestamp),
		}
	}
	var restrictions []orm.GlobalSkinRestriction
	if err := orm.GormDB.Order("skin_id asc").Find(&restrictions).Error; err != nil {
		return 0, 12201, err
	}

	response.ForbiddenSkinList = make([]uint32, 0, len(restrictions))
	response.ForbiddenSkinType = make([]uint32, 0, len(restrictions))
	for _, restriction := range restrictions {
		response.ForbiddenSkinList = append(response.ForbiddenSkinList, restriction.SkinID)
		response.ForbiddenSkinType = append(response.ForbiddenSkinType, restriction.Type)
	}

	var windows []orm.GlobalSkinRestrictionWindow
	if err := orm.GormDB.Order("skin_id asc").Find(&windows).Error; err != nil {
		return 0, 12201, err
	}
	response.ForbiddenList = make([]*protobuf.SKIN_FORBIDDEN, 0, len(windows))
	for _, window := range windows {
		response.ForbiddenList = append(response.ForbiddenList, &protobuf.SKIN_FORBIDDEN{
			Id:        proto.Uint32(window.SkinID),
			Type:      proto.Uint32(window.Type),
			StartTime: proto.Uint32(window.StartTime),
			StopTime:  proto.Uint32(window.StopTime),
		})
	}

	return client.SendMessage(12201, &response)
}
