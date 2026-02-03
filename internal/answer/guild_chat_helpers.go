package answer

import (
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

const guildChatPlaceholderID uint32 = 0

func buildGuildChatPlayer(commander *orm.Commander) *protobuf.PLAYER_INFO_P60 {
	return &protobuf.PLAYER_INFO_P60{
		Id:   proto.Uint32(commander.CommanderID),
		Name: proto.String(commander.Name),
		Lv:   proto.Uint32(uint32(commander.Level)),
		Display: &protobuf.DISPLAYINFO{
			Icon:          proto.Uint32(commander.DisplayIconID),
			Skin:          proto.Uint32(commander.DisplaySkinID),
			IconFrame:     proto.Uint32(commander.SelectedIconFrameID),
			ChatFrame:     proto.Uint32(commander.SelectedChatFrameID),
			IconTheme:     proto.Uint32(commander.DisplayIconThemeID),
			MarryFlag:     proto.Uint32(0),
			TransformFlag: proto.Uint32(0),
		},
	}
}
