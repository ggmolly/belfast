package answer

import (
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func PlayerInfo(buffer *[]byte, client *connection.Client) (int, int, error) {
	if err := ensureGuideIndices(client.Commander); err != nil {
		return 0, 11003, err
	}

	response := protobuf.SC_11003{
		Id:                 proto.Uint32(uint32(client.Commander.CommanderID)),
		Name:               proto.String(client.Commander.Name),
		Level:              proto.Uint32(uint32(client.Commander.Level)),
		Exp:                proto.Uint32(uint32(client.Commander.Exp)),
		ChildDisplay:       proto.Uint32(1004),
		AttackCount:        proto.Uint32(0),
		WinCount:           proto.Uint32(0),
		Adv:                proto.String(client.Commander.Manifesto),
		ShipBagMax:         proto.Uint32(250),
		EquipBagMax:        proto.Uint32(250),
		GmFlag:             proto.Uint32(0),
		Rank:               proto.Uint32(0),
		PvpAttackCount:     proto.Uint32(0),
		PvpWinCount:        proto.Uint32(0),
		CollectAttackCount: proto.Uint32(client.Commander.CollectAttackCount),
		GuideIndex:         proto.Uint32(client.Commander.GuideIndex),
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
			Icon:          proto.Uint32(client.Commander.DisplayIconID),
			Skin:          proto.Uint32(client.Commander.DisplaySkinID),
			IconFrame:     proto.Uint32(client.Commander.SelectedIconFrameID),
			ChatFrame:     proto.Uint32(client.Commander.SelectedChatFrameID),
			IconTheme:     proto.Uint32(client.Commander.DisplayIconThemeID),
			MarryFlag:     proto.Uint32(0),
			TransformFlag: proto.Uint32(0),
		},
		Rmb: proto.Uint32(999), // No idea what this is
		Appreciation: &protobuf.APPRECIATIONINFO{
			MusicNo:   proto.Uint32(0),
			MusicMode: proto.Uint32(0),
		},
		ThemeUploadNotAllowedTime: proto.Uint32(0),
		RandomShipMode:            proto.Uint32(client.Commander.RandomShipMode),
		MarryShip:                 proto.Uint32(0),
		Cover: &protobuf.LIVINGAREA_COVER{
			Id: proto.Uint32(client.Commander.LivingAreaCoverID),
		},
		MailStoreroomLv: proto.Uint32(1),
		BattleUi:        proto.Uint32(client.Commander.SelectedBattleUIID),
		NewGuideIndex:   proto.Uint32(client.Commander.NewGuideIndex),
	}

	// Get user's secretaries
	secretaries := client.Commander.GetSecretaries()
	response.Character = make([]*protobuf.KVDATA, len(secretaries))
	for i, secretary := range secretaries {
		response.Character[i] = &protobuf.KVDATA{
			Key:   proto.Uint32(uint32(secretary.ID)),
			Value: proto.Uint32(secretary.SecretaryPhantomID),
		}
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
	if response.Display.GetIcon() == 0 && len(secretaries) > 0 {
		response.Display.Icon = proto.Uint32(secretaries[0].ShipID)
	}
	if response.Display.GetSkin() == 0 && len(secretaries) > 0 {
		response.Display.Skin = proto.Uint32(secretaries[0].SkinID)
	}
	flags, err := orm.ListCommanderCommonFlags(client.Commander.CommanderID)
	if err != nil {
		return 0, 11003, err
	}
	response.FlagList = flags
	storyIDs, err := orm.ListCommanderStoryIDs(client.Commander.CommanderID)
	if err != nil {
		return 0, 11003, err
	}
	response.StoryList = storyIDs
	soundStoryIDs, err := orm.ListCommanderSoundStoryIDs(client.Commander.CommanderID)
	if err != nil {
		return 0, 11003, err
	}
	response.Soundstory = soundStoryIDs
	medalIDs, err := orm.ListCommanderMedalDisplay(client.Commander.CommanderID)
	if err != nil {
		return 0, 11003, err
	}
	response.MedalId = medalIDs
	appreciationState, err := orm.GetOrCreateCommanderAppreciationState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		return 0, 11003, err
	}
	response.Appreciation.MusicNo = proto.Uint32(appreciationState.MusicNo)
	response.Appreciation.MusicMode = proto.Uint32(appreciationState.MusicMode)
	response.CartoonReadMark = orm.ToUint32List(appreciationState.CartoonReadMark)
	response.CartoonCollectMark = orm.ToUint32List(appreciationState.CartoonCollectMark)
	response.Appreciation.Gallerys = orm.ToUint32List(appreciationState.GalleryUnlocks)
	response.Appreciation.FavorGallerys = orm.ToUint32List(appreciationState.GalleryFavorIds)
	response.Appreciation.FavorMusics = orm.ToUint32List(appreciationState.MusicFavorIds)
	attires, err := orm.ListCommanderAttires(client.Commander.CommanderID)
	if err != nil {
		return 0, 11003, err
	}
	now := time.Now()
	iconFrames := make([]*protobuf.IDTIMEINFO, 0)
	chatFrames := make([]*protobuf.IDTIMEINFO, 0)
	battleUI := make([]uint32, 0)
	for _, attire := range attires {
		if attire.ExpiresAt != nil && attire.ExpiresAt.Before(now) {
			continue
		}
		var expires uint32
		if attire.ExpiresAt != nil {
			expires = uint32(attire.ExpiresAt.Unix())
		}
		info := &protobuf.IDTIMEINFO{
			Id:   proto.Uint32(attire.AttireID),
			Time: proto.Uint32(expires),
		}
		switch attire.Type {
		case consts.AttireTypeIconFrame:
			iconFrames = append(iconFrames, info)
		case consts.AttireTypeChatFrame:
			chatFrames = append(chatFrames, info)
		case consts.AttireTypeCombatUI:
			battleUI = append(battleUI, attire.AttireID)
		}
	}
	if !containsUint32(battleUI, 0) {
		battleUI = append([]uint32{0}, battleUI...)
	}
	response.IconFrameList = iconFrames
	response.ChatFrameList = chatFrames
	response.BattleUiList = battleUI
	if client.Commander.SelectedBattleUIID != 0 && !containsUint32(battleUI, client.Commander.SelectedBattleUIID) {
		response.BattleUiList = append(response.BattleUiList, client.Commander.SelectedBattleUIID)
	}
	coverEntries, err := orm.ListCommanderLivingAreaCovers(client.Commander.CommanderID)
	if err != nil {
		return 0, 11003, err
	}
	coverIDs := make([]uint32, 0, len(coverEntries)+1)
	if client.Commander.LivingAreaCoverID == 0 {
		coverIDs = append(coverIDs, 0)
	}
	for _, cover := range coverEntries {
		coverIDs = append(coverIDs, cover.CoverID)
	}
	if client.Commander.LivingAreaCoverID != 0 && !containsUint32(coverIDs, client.Commander.LivingAreaCoverID) {
		coverIDs = append(coverIDs, client.Commander.LivingAreaCoverID)
	}
	response.Cover.Covers = coverIDs
	if !client.Commander.NameChangeCooldown.IsZero() && !client.Commander.NameChangeCooldown.Equal(time.Unix(0, 0)) {
		response.CdList = append(response.CdList, &protobuf.COOLDOWN{
			Key:       proto.Uint32(1),
			Timestamp: proto.Uint32(uint32(client.Commander.NameChangeCooldown.Unix())),
		})
	}

	response.ChatRoomId = proto.Uint32(client.Commander.RoomID)
	bytesWritten, packetId, err := client.SendMessage(11003, &response)
	if err != nil {
		return bytesWritten, packetId, err
	}
	// TODO: Remove this once the client explicitly requests SC_22300 on login.
	manualBuffer := []byte{}
	if _, _, err := CommanderManualInfo(&manualBuffer, client); err != nil {
		return bytesWritten, packetId, err
	}
	return bytesWritten, packetId, nil
}

func ensureGuideIndices(commander *orm.Commander) error {
	updates := make(map[string]any)
	if commander.GuideIndex == 0 {
		updates["guide_index"] = uint32(1)
		commander.GuideIndex = 1
	}
	if commander.NewGuideIndex == 0 {
		updates["new_guide_index"] = uint32(1)
		commander.NewGuideIndex = 1
	}
	if len(updates) == 0 {
		return nil
	}
	// TODO: Align guide index backfill with guide versioning rules.
	return orm.GormDB.Model(&orm.Commander{}).
		Where("commander_id = ?", commander.CommanderID).
		Updates(updates).Error
}

func containsUint32(list []uint32, value uint32) bool {
	for _, entry := range list {
		if entry == value {
			return true
		}
	}
	return false
}
