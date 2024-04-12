package answer

import (
	"github.com/ggmolly/belfast/connection"
	"github.com/ggmolly/belfast/logger"

	"github.com/ggmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

// GuideIndex represents the game's tutorial progress
// 0 = breaks
// 1 = guides through the main UI		sends CS_11017 -> UpdateStoryCommand
// 2 = nothing apparent		 			sends CS_11017
// 29 = works, idk what it does, but doesn't break anything

func PlayerInfo(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_11003{
		Id:                 proto.Uint32(uint32(client.Commander.CommanderID)),
		Name:               proto.String(client.Commander.Name),
		Level:              proto.Uint32(uint32(client.Commander.Level)),
		Exp:                proto.Uint32(uint32(client.Commander.Exp)),
		ChildDisplay:       proto.Uint32(1004),
		AttackCount:        proto.Uint32(0),
		WinCount:           proto.Uint32(0),
		Adv:                proto.String(""),
		ShipBagMax:         proto.Uint32(250),
		EquipBagMax:        proto.Uint32(250),
		GmFlag:             proto.Uint32(0),
		Rank:               proto.Uint32(0),
		PvpAttackCount:     proto.Uint32(0),
		PvpWinCount:        proto.Uint32(0),
		CollectAttackCount: proto.Uint32(0),
		GuideIndex:         proto.Uint32(29), // See notes above
		BuyOilCount:        proto.Uint32(0),
		ChatRoomId:         proto.Uint32(0),
		MaxRank:            proto.Uint32(0),
		RegisterTime:       proto.Uint32(0),
		ShipCount:          proto.Uint32(uint32(len(client.Commander.Ships))),
		AccPayLv:           proto.Uint32(0),
		GuildWaitTime:      proto.Uint32(0),
		ChatMsgBanTime:     proto.Uint32(0),
		CommanderBagMax:    proto.Uint32(250),
		Display: &protobuf.DISPLAYINFO{
			Icon:          proto.Uint32(202124), // Should display Belfast's icon
			Skin:          proto.Uint32(202123), // Should display Belfast's default skin
			IconFrame:     proto.Uint32(0),
			ChatFrame:     proto.Uint32(0),
			IconTheme:     proto.Uint32(0),
			MarryFlag:     proto.Uint32(0),
			TransformFlag: proto.Uint32(0),
		},
		Rmb:                       proto.Uint32(999), // No idea what this is
		Appreciation:              &protobuf.APPRECIATIONINFO{},
		ThemeUploadNotAllowedTime: proto.Uint32(0),
		RandomShipMode:            proto.Uint32(0),
		MarryShip:                 proto.Uint32(0),
	}

	// Get user's secretaries
	secretaries := client.Commander.GetSecretaries()
	response.Character = make([]uint32, len(secretaries))
	for i, secretary := range secretaries {
		response.Character[i] = uint32(secretary.ID)
	}
	if len(response.Character) == 0 {
		logger.LogEvent("Server", "PlayerInfo", "No secretaries found", logger.LOG_LEVEL_ERROR)
		return 0, 11003, nil
	}

	response.ResourceList = make([]*protobuf.RESOURCE, len(client.Commander.OwnedResources))
	for i, resource := range client.Commander.OwnedResources {
		response.ResourceList[i] = &protobuf.RESOURCE{
			Type: proto.Uint32(resource.ResourceID),
			Num:  proto.Uint32(resource.Amount),
		}
	}

	response.ChatRoomId = proto.Uint32(client.Commander.RoomID)
	return client.SendMessage(11003, &response)
}
