package entrypoint

import (
	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/packets"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

var validRegions = map[string]interface{}{
	"CN": nil,
	"EN": nil,
	"JP": nil,
	"KR": nil,
	"TW": nil,
}

func registerPackets() {
	packets.RegisterPacketHandler(10800, []packets.PacketHandler{answer.Forge_SC10801})
	packets.RegisterPacketHandler(10700, []packets.PacketHandler{answer.GatewayPackInfo})
	packets.RegisterPacketHandler(8239, []packets.PacketHandler{answer.Forge_SC8239})
	packets.RegisterPacketHandler(10020, []packets.PacketHandler{answer.Forge_SC10021})
	packets.RegisterLocalizedPacketHandler(10802, packets.LocalizedHandler{
		CN: &[]packets.PacketHandler{answer.Forge_SC10803_CN_JP_KR_TW},
		TW: &[]packets.PacketHandler{answer.Forge_SC10803_CN_JP_KR_TW},
		JP: &[]packets.PacketHandler{answer.Forge_SC10803_CN_JP_KR_TW},
		KR: &[]packets.PacketHandler{answer.Forge_SC10803_CN_JP_KR_TW},
	})
	packets.RegisterPacketHandler(10018, []packets.PacketHandler{answer.Forge_SC10019})
	packets.RegisterPacketHandler(10022, []packets.PacketHandler{answer.JoinServer})
	packets.RegisterPacketHandler(10024, []packets.PacketHandler{answer.CreateNewPlayer})
	packets.RegisterPacketHandler(10026, []packets.PacketHandler{answer.PlayerExist})
	packets.RegisterPacketHandler(11001, []packets.PacketHandler{
		answer.LastLogin,
		answer.PlayerInfo,
		answer.PlayerBuffs,
		answer.GetMetaProgress,
		answer.LastOnlineInfo,
		answer.ResourcesInfo,
		answer.EventData,
		answer.Meowfficers,
		answer.CommanderCollection,
		answer.OngoingBuilds,
		answer.PlayerDock,
		answer.CommanderDock,
		answer.CommanderFleet,
		answer.CommanderOwnedSkins,
		answer.TechnologyRefreshList,
		answer.ShipyardData,
		answer.TechnologyNationProxy,
		answer.CommanderStoryProgress,
		answer.EventCollectionInfo,
		answer.CommanderCommissionsFleet,
		answer.ShopData,
		answer.WorldBaseInfo,
		answer.ChapterBaseSync,
		answer.EquipedSpecialWeapons,
		answer.EquippedWeaponSkin,
		answer.OwnedItems,
		answer.CommanderMissions,
		answer.WeeklyMissions,
		answer.DormData,
		answer.FleetEnergyRecoverTime,
		answer.GameMailbox,
		answer.CompensateNotification,
		answer.CommanderFriendList,
		answer.Activities,
		answer.PermanentActivites,
		answer.GameNotices,
		answer.SendPlayerShipCount,
	})
	packets.RegisterPacketHandler(25026, []packets.PacketHandler{answer.GetCommanderHome})
	packets.RegisterPacketHandler(34501, []packets.PacketHandler{answer.WorldBossInfo})
	packets.RegisterPacketHandler(63317, []packets.PacketHandler{answer.MetaCharacterTacticsInfoRequestCommandResponse})
	packets.RegisterPacketHandler(34001, []packets.PacketHandler{answer.GetMetaShipsPointsResponse})
	packets.RegisterPacketHandler(18001, []packets.PacketHandler{answer.ExerciseEnemies})
	packets.RegisterPacketHandler(18003, []packets.PacketHandler{answer.ExerciseReplaceRivals})
	packets.RegisterPacketHandler(18006, []packets.PacketHandler{answer.ExercisePowerRankList})
	packets.RegisterPacketHandler(18008, []packets.PacketHandler{answer.UpdateExerciseFleet})
	packets.RegisterPacketHandler(18100, []packets.PacketHandler{answer.GetArenaShop})
	packets.RegisterPacketHandler(18102, []packets.PacketHandler{answer.RefreshArenaShop})
	packets.RegisterPacketHandler(18104, []packets.PacketHandler{answer.GetRivalInfo})
	packets.RegisterPacketHandler(18201, []packets.PacketHandler{answer.BillboardRankListPage})
	packets.RegisterPacketHandler(18203, []packets.PacketHandler{answer.BillboardMyRank})
	packets.RegisterPacketHandler(60033, []packets.PacketHandler{answer.GetGuildShop})
	packets.RegisterPacketHandler(16106, []packets.PacketHandler{answer.GetMedalShop})
	packets.RegisterPacketHandler(16108, []packets.PacketHandler{answer.MedalShopPurchase})
	packets.RegisterPacketHandler(60037, []packets.PacketHandler{answer.CommanderGuildData})
	packets.RegisterPacketHandler(26150, []packets.PacketHandler{answer.GetMiniGameShop})
	packets.RegisterPacketHandler(11506, []packets.PacketHandler{answer.ClickMingShi})
	packets.RegisterPacketHandler(62100, []packets.PacketHandler{answer.CommanderGuildTechnologies})
	packets.RegisterPacketHandler(26101, []packets.PacketHandler{answer.MiniGameHubData})
	packets.RegisterPacketHandler(24020, []packets.PacketHandler{answer.LimitChallengeInfo})
	packets.RegisterPacketHandler(24004, []packets.PacketHandler{answer.ChallengeInfo})
	packets.RegisterPacketHandler(26051, []packets.PacketHandler{answer.AtelierRequest})
	packets.RegisterPacketHandler(11601, []packets.PacketHandler{answer.EmojiInfoRequest})
	packets.RegisterPacketHandler(11603, []packets.PacketHandler{answer.FetchSecondaryPasswordCommandResponse})
	packets.RegisterPacketHandler(11605, []packets.PacketHandler{answer.SetSecondaryPasswordCommandResponse})
	packets.RegisterPacketHandler(11607, []packets.PacketHandler{answer.SetSecondaryPasswordSettingsCommandResponse})
	packets.RegisterPacketHandler(11609, []packets.PacketHandler{answer.ConfirmSecondaryPasswordCommandResponse})
	packets.RegisterPacketHandler(17005, []packets.PacketHandler{answer.CollectionGetAward17005})
	packets.RegisterPacketHandler(17201, []packets.PacketHandler{answer.FetchVoteTicketInfo})
	packets.RegisterPacketHandler(17203, []packets.PacketHandler{answer.FetchVoteInfo})
	packets.RegisterPacketHandler(16104, []packets.PacketHandler{answer.GetChargeList})
	packets.RegisterPacketHandler(60100, []packets.PacketHandler{answer.CommanderGuildChat})
	packets.RegisterPacketHandler(60007, []packets.PacketHandler{answer.GuildSendMessage})
	packets.RegisterPacketHandler(60102, []packets.PacketHandler{answer.GuildGetUserInfoCommand})
	packets.RegisterPacketHandler(61009, []packets.PacketHandler{answer.GetMyAssaultFleetCommandResponse})
	packets.RegisterPacketHandler(61011, []packets.PacketHandler{answer.GuildGetAssaultFleetCommandResponse})
	packets.RegisterPacketHandler(61005, []packets.PacketHandler{answer.GuildGetActivationEventCommandResponse})
	packets.RegisterPacketHandler(60003, []packets.PacketHandler{answer.GetGuildRequestsCommandResponse})
	packets.RegisterPacketHandler(13501, []packets.PacketHandler{answer.RemasterSetActiveChapter})
	packets.RegisterPacketHandler(13503, []packets.PacketHandler{answer.RemasterTickets})
	packets.RegisterPacketHandler(13505, []packets.PacketHandler{answer.RemasterInfo})
	packets.RegisterPacketHandler(13507, []packets.PacketHandler{answer.RemasterAwardReceive})
	packets.RegisterPacketHandler(13301, []packets.PacketHandler{answer.EscortQuery})
	packets.RegisterPacketHandler(13401, []packets.PacketHandler{answer.GetSubmarineExpeditionInfo})
	packets.RegisterPacketHandler(13403, []packets.PacketHandler{answer.SubmarineChapterInfo})
	packets.RegisterPacketHandler(11202, []packets.PacketHandler{answer.ActivityOperation})
	packets.RegisterPacketHandler(11204, []packets.PacketHandler{answer.EditActivityFleet})
	packets.RegisterPacketHandler(11206, []packets.PacketHandler{answer.ActivityPermanentStart})
	packets.RegisterPacketHandler(11208, []packets.PacketHandler{answer.ActivityPermanentFinish})
	packets.RegisterPacketHandler(13003, []packets.PacketHandler{answer.EventCollectionStart})
	packets.RegisterPacketHandler(13007, []packets.PacketHandler{answer.EventGiveUp})
	packets.RegisterPacketHandler(13009, []packets.PacketHandler{answer.EventFlush})
	packets.RegisterPacketHandler(13005, []packets.PacketHandler{answer.EventFinish})
	packets.RegisterPacketHandler(11751, []packets.PacketHandler{answer.RefluxRequestData})
	packets.RegisterPacketHandler(11722, []packets.PacketHandler{answer.InstagramChatActivateTopic})
	packets.RegisterPacketHandler(11005, []packets.PacketHandler{answer.AttireApply})
	packets.RegisterPacketHandler(11007, []packets.PacketHandler{answer.ChangePlayerName})
	packets.RegisterPacketHandler(11009, []packets.PacketHandler{answer.ChangeManifesto})
	packets.RegisterPacketHandler(11016, []packets.PacketHandler{answer.UpdateGuideIndex})
	packets.RegisterPacketHandler(11017, []packets.PacketHandler{answer.UpdateStory})
	packets.RegisterPacketHandler(11025, []packets.PacketHandler{answer.SurveyRequest})
	packets.RegisterPacketHandler(11027, []packets.PacketHandler{answer.SurveyState})
	packets.RegisterPacketHandler(11030, []packets.PacketHandler{answer.ChangeLivingAreaCover})
	packets.RegisterPacketHandler(11032, []packets.PacketHandler{answer.UpdateStoryList})
	packets.RegisterPacketHandler(10100, []packets.PacketHandler{answer.SendHeartbeat})
	packets.RegisterPacketHandler(11100, []packets.PacketHandler{answer.SendCmd})
	packets.RegisterPacketHandler(11013, []packets.PacketHandler{answer.GiveResources})
	packets.RegisterPacketHandler(15002, []packets.PacketHandler{answer.UseItem})
	packets.RegisterPacketHandler(15004, []packets.PacketHandler{answer.ItemOp15004})
	packets.RegisterPacketHandler(15006, []packets.PacketHandler{answer.ComposeItem})
	packets.RegisterPacketHandler(15012, []packets.PacketHandler{answer.QuickExchangeBlueprint})
	packets.RegisterPacketHandler(15008, []packets.PacketHandler{answer.SellItem})
	packets.RegisterPacketHandler(15010, []packets.PacketHandler{answer.ProposeExchangeRing})
	packets.RegisterPacketHandler(33000, []packets.PacketHandler{answer.WorldCheckInfo})
	packets.RegisterPacketHandler(10994, []packets.PacketHandler{answer.CheaterMark})
	packets.RegisterPacketHandler(10996, []packets.PacketHandler{answer.VersionCheck})
	packets.RegisterPacketHandler(29001, []packets.PacketHandler{answer.NewEducateRequest})
	packets.RegisterPacketHandler(29003, []packets.PacketHandler{answer.NewEducateGetEndings})
	packets.RegisterPacketHandler(29005, []packets.PacketHandler{answer.NewEducateSelectEnding})
	packets.RegisterPacketHandler(29007, []packets.PacketHandler{answer.NewEducateReset})
	packets.RegisterPacketHandler(29009, []packets.PacketHandler{answer.NewEducateSetCall})
	packets.RegisterPacketHandler(29011, []packets.PacketHandler{answer.NewEducateMainEvent})
	packets.RegisterPacketHandler(29013, []packets.PacketHandler{answer.NewEducateAssess})
	packets.RegisterPacketHandler(29015, []packets.PacketHandler{answer.NewEducateGetTopics})
	packets.RegisterPacketHandler(29017, []packets.PacketHandler{answer.NewEducateSelectTopic})
	packets.RegisterPacketHandler(29019, []packets.PacketHandler{answer.NewEducateGetTalents})
	packets.RegisterPacketHandler(29021, []packets.PacketHandler{answer.NewEducateRefreshTalent})
	packets.RegisterPacketHandler(29023, []packets.PacketHandler{answer.NewEducateSelectTalent})
	packets.RegisterPacketHandler(29025, []packets.PacketHandler{answer.NewEducateChangePhase})
	packets.RegisterPacketHandler(29027, []packets.PacketHandler{answer.NewEducateUpgradeFavor})
	packets.RegisterPacketHandler(29030, []packets.PacketHandler{answer.NewEducateTriggerNode})
	packets.RegisterPacketHandler(29032, []packets.PacketHandler{answer.NewEducateClearNodeChain})
	packets.RegisterPacketHandler(29040, []packets.PacketHandler{answer.NewEducateSchedule})
	packets.RegisterPacketHandler(29042, []packets.PacketHandler{answer.NewEducateNextPlan})
	packets.RegisterPacketHandler(29044, []packets.PacketHandler{answer.NewEducateUpgradePlan})
	packets.RegisterPacketHandler(29046, []packets.PacketHandler{answer.NewEducateScheduleSkip})
	packets.RegisterPacketHandler(29048, []packets.PacketHandler{answer.NewEducateGetExtraDrop})
	packets.RegisterPacketHandler(29060, []packets.PacketHandler{answer.NewEducateGetMap})
	packets.RegisterPacketHandler(29062, []packets.PacketHandler{answer.NewEducateMapNormal})
	packets.RegisterPacketHandler(29064, []packets.PacketHandler{answer.NewEducateMapEvent})
	packets.RegisterPacketHandler(29066, []packets.PacketHandler{answer.NewEducateShopping})
	packets.RegisterPacketHandler(29068, []packets.PacketHandler{answer.NewEducateMapShip})
	packets.RegisterPacketHandler(29070, []packets.PacketHandler{answer.NewEducateUpgradeNormalSite})
	packets.RegisterPacketHandler(29090, []packets.PacketHandler{answer.NewEducateSelectMind})
	packets.RegisterPacketHandler(29092, []packets.PacketHandler{answer.NewEducateRefresh})
	packets.RegisterPacketHandler(30101, []packets.PacketHandler{answer.CompensateNotification})
	packets.RegisterPacketHandler(28000, []packets.PacketHandler{answer.Dorm3dApartmentData})
	packets.RegisterPacketHandler(28026, []packets.PacketHandler{answer.Dorm3dInstagramOp})
	packets.RegisterPacketHandler(28028, []packets.PacketHandler{answer.Dorm3dInstagramDiscuss})
	packets.RegisterPacketHandler(12002, []packets.PacketHandler{answer.ShipBuild})
	packets.RegisterPacketHandler(12008, []packets.PacketHandler{answer.BuildQuickFinish})
	packets.RegisterPacketHandler(12011, []packets.PacketHandler{answer.RemouldShip})
	packets.RegisterPacketHandler(12043, []packets.PacketHandler{answer.BuildFinish})
	packets.RegisterPacketHandler(12020, []packets.PacketHandler{answer.ShipAction12020})
	packets.RegisterPacketHandler(12025, []packets.PacketHandler{answer.GetShip})
	packets.RegisterPacketHandler(12027, []packets.PacketHandler{answer.UpgradeStar})
	packets.RegisterPacketHandler(12029, []packets.PacketHandler{answer.ShipAction12029})
	packets.RegisterPacketHandler(12045, []packets.PacketHandler{answer.ConfirmShip})
	packets.RegisterPacketHandler(12006, []packets.PacketHandler{answer.EquipToShip})
	packets.RegisterPacketHandler(12036, []packets.PacketHandler{answer.UpdateShipEquipmentSkin})
	packets.RegisterPacketHandler(16100, []packets.PacketHandler{answer.SupportShipRequisition})
	packets.RegisterPacketHandler(12047, []packets.PacketHandler{answer.ExchangeShip})
	packets.RegisterPacketHandler(30002, []packets.PacketHandler{answer.SendMailList})
	packets.RegisterPacketHandler(30004, []packets.PacketHandler{answer.GetCollectionMailList})
	packets.RegisterPacketHandler(30006, []packets.PacketHandler{answer.HandleMailDealCmd})
	packets.RegisterPacketHandler(30008, []packets.PacketHandler{answer.DeleteArchivedMail})
	packets.RegisterPacketHandler(30102, []packets.PacketHandler{answer.GetCompensateList})
	packets.RegisterPacketHandler(30104, []packets.PacketHandler{answer.GetCompensateReward})
	packets.RegisterPacketHandler(22300, []packets.PacketHandler{answer.CommanderManualInfo})
	packets.RegisterPacketHandler(22302, []packets.PacketHandler{answer.CommanderManualGetTask})
	packets.RegisterPacketHandler(22304, []packets.PacketHandler{answer.CommanderManualGetPtAward})
	packets.RegisterPacketHandler(11701, []packets.PacketHandler{answer.JuustagramOp})
	packets.RegisterPacketHandler(11703, []packets.PacketHandler{answer.JuustagramComment})
	packets.RegisterPacketHandler(11705, []packets.PacketHandler{answer.JuustagramMessageRange})
	packets.RegisterPacketHandler(11710, []packets.PacketHandler{answer.JuustagramData})
	packets.RegisterPacketHandler(11712, []packets.PacketHandler{answer.InstagramChatReply})
	packets.RegisterPacketHandler(11714, []packets.PacketHandler{answer.InstagramChatSetSkin})
	packets.RegisterPacketHandler(11716, []packets.PacketHandler{answer.InstagramChatSetCare})
	packets.RegisterPacketHandler(11718, []packets.PacketHandler{answer.InstagramChatSetTopic})
	packets.RegisterPacketHandler(11720, []packets.PacketHandler{answer.JuustagramReadTip})
	packets.RegisterPacketHandler(11753, []packets.PacketHandler{answer.RefluxSign})
	packets.RegisterPacketHandler(11755, []packets.PacketHandler{answer.RefluxGetPTAward})
	packets.RegisterPacketHandler(11800, []packets.PacketHandler{answer.GetShipCount})
	packets.RegisterPacketHandler(10991, []packets.PacketHandler{answer.GameTracking})
	packets.RegisterPacketHandler(10992, []packets.PacketHandler{answer.NewTracking})
	packets.RegisterPacketHandler(10993, []packets.PacketHandler{answer.TrackCommand})
	packets.RegisterPacketHandler(11212, []packets.PacketHandler{answer.UrExchangeTracking})
	packets.RegisterPacketHandler(11029, []packets.PacketHandler{answer.MainSceneTracking})
	packets.RegisterPacketHandler(11023, []packets.PacketHandler{answer.GetRefundInfo})
	packets.RegisterPacketHandler(11025, []packets.PacketHandler{answer.SurveyRequest})
	packets.RegisterPacketHandler(11027, []packets.PacketHandler{answer.SurveyState})
	packets.RegisterPacketHandler(22101, []packets.PacketHandler{answer.GetShopStreet})
	packets.RegisterPacketHandler(16001, []packets.PacketHandler{answer.ShoppingCommandAnswer})
	packets.RegisterPacketHandler(16201, []packets.PacketHandler{answer.MonthShopPurchase})
	packets.RegisterPacketHandler(16203, []packets.PacketHandler{answer.MonthShopFlag})
	packets.RegisterPacketHandler(16205, []packets.PacketHandler{answer.CryptolaliaUnlock})
	packets.RegisterPacketHandler(11501, []packets.PacketHandler{answer.ChargeCommandAnswer})
	packets.RegisterPacketHandler(11504, []packets.PacketHandler{answer.ChargeConfirmCommandAnswer})
	packets.RegisterPacketHandler(11508, []packets.PacketHandler{answer.ExchangeCodeRedeem})
	packets.RegisterPacketHandler(11510, []packets.PacketHandler{answer.ChargeFailedCommandAnswer})
	packets.RegisterPacketHandler(11513, []packets.PacketHandler{answer.RefundChargeCommandAnswer})
	packets.RegisterPacketHandler(12004, []packets.PacketHandler{answer.RetireShip})
	packets.RegisterPacketHandler(12017, []packets.PacketHandler{answer.ModShip})
	packets.RegisterPacketHandler(11401, []packets.PacketHandler{answer.ChatRoomChange})
	packets.RegisterPacketHandler(50102, []packets.PacketHandler{answer.ReceiveChatMessage})
	packets.RegisterPacketHandler(12032, []packets.PacketHandler{answer.ProposeShip})
	packets.RegisterPacketHandler(20007, []packets.PacketHandler{func(b *[]byte, c *connection.Client) (int, int, error) {
		response := protobuf.SC_20008{
			Result: proto.Uint32(1),
		}
		return c.SendMessage(20008, &response)
	}})
	packets.RegisterPacketHandler(11011, []packets.PacketHandler{answer.UpdateSecretaries})
	packets.RegisterPacketHandler(12038, []packets.PacketHandler{answer.UpgradeShipMaxLevel})
	packets.RegisterPacketHandler(12040, []packets.PacketHandler{answer.SetFavoriteShip})
	packets.RegisterPacketHandler(12022, []packets.PacketHandler{answer.ChangeShipLockState})
	packets.RegisterPacketHandler(12202, []packets.PacketHandler{answer.ChangeSelectedSkin})
	packets.RegisterPacketHandler(12204, []packets.PacketHandler{answer.ToggleRandomFlagShip})
	packets.RegisterPacketHandler(12206, []packets.PacketHandler{answer.ChangeRandomFlagShipMode})
	packets.RegisterPacketHandler(12208, []packets.PacketHandler{answer.ChangeRandomFlagShips})
	packets.RegisterPacketHandler(12210, []packets.PacketHandler{answer.FinishPhantomQuest})
	packets.RegisterPacketHandler(12212, []packets.PacketHandler{answer.GetPhantomQuestProgress})
	packets.RegisterPacketHandler(14002, []packets.PacketHandler{answer.UpgradeEquipmentOnShip14002})
	packets.RegisterPacketHandler(14006, []packets.PacketHandler{answer.CompositeEquipment})
	packets.RegisterPacketHandler(14004, []packets.PacketHandler{answer.UpgradeEquipmentInBag14004})
	packets.RegisterPacketHandler(14008, []packets.PacketHandler{answer.DestroyEquipments})
	packets.RegisterPacketHandler(14010, []packets.PacketHandler{answer.RevertEquipment})
	packets.RegisterPacketHandler(14013, []packets.PacketHandler{answer.TransformEquipmentOnShip14013})
	packets.RegisterPacketHandler(14015, []packets.PacketHandler{answer.TransformEquipmentInBag14015})
	packets.RegisterPacketHandler(14201, []packets.PacketHandler{answer.EquipSpWeapon})
	packets.RegisterPacketHandler(14203, []packets.PacketHandler{answer.UpgradeSpWeapon})
	packets.RegisterPacketHandler(14205, []packets.PacketHandler{answer.ReforgeSpWeapon})
	packets.RegisterPacketHandler(14207, []packets.PacketHandler{answer.ConfirmReforgeSpWeapon})
	packets.RegisterPacketHandler(14209, []packets.PacketHandler{answer.CompositeSpWeapon})
	packets.RegisterPacketHandler(12301, []packets.PacketHandler{answer.ReqPlayerAssistShip})
	packets.RegisterPacketHandler(12034, []packets.PacketHandler{answer.RenameProposedShip})
	packets.RegisterPacketHandler(27000, []packets.PacketHandler{answer.EducateRequest})
	packets.RegisterPacketHandler(27010, []packets.PacketHandler{func(b *[]byte, c *connection.Client) (int, int, error) {
		response := protobuf.SC_27011{}
		return c.SendMessage(27011, &response)
	}})
	packets.RegisterPacketHandler(12102, []packets.PacketHandler{answer.FleetCommit})
	packets.RegisterPacketHandler(12104, []packets.PacketHandler{answer.FleetRename})
	packets.RegisterLocalizedPacketHandler(13101, packets.LocalizedHandler{
		CN: &[]packets.PacketHandler{answer.ChapterTracking},
		EN: &[]packets.PacketHandler{answer.ChapterTracking},
		JP: &[]packets.PacketHandler{answer.ChapterTracking},
		KR: &[]packets.PacketHandler{answer.ChapterTrackingKR},
		TW: &[]packets.PacketHandler{answer.ChapterTracking},
	})
	packets.RegisterPacketHandler(13103, []packets.PacketHandler{answer.ChapterOp})
	packets.RegisterPacketHandler(13109, []packets.PacketHandler{answer.GetChapterDropShipList})
	packets.RegisterPacketHandler(13106, []packets.PacketHandler{answer.ChapterBattleResultRequest})
	packets.RegisterPacketHandler(13107, []packets.PacketHandler{func(b *[]byte, c *connection.Client) (int, int, error) {
		response := protobuf.SC_13108{
			Result: proto.Uint32(0),
		}
		return c.SendMessage(13108, &response)
	}})
	packets.RegisterPacketHandler(13111, []packets.PacketHandler{answer.RemoveEliteTargetShip})
	packets.RegisterPacketHandler(40001, []packets.PacketHandler{answer.BeginStage})
	packets.RegisterPacketHandler(40003, []packets.PacketHandler{answer.FinishStage})
	packets.RegisterPacketHandler(40005, []packets.PacketHandler{answer.QuitBattle})
	packets.RegisterPacketHandler(40007, []packets.PacketHandler{answer.DailyQuickBattle})
	packets.RegisterPacketHandler(11019, []packets.PacketHandler{answer.UpdateCommonFlagCommand})
	packets.RegisterPacketHandler(11021, []packets.PacketHandler{answer.CancelCommonFlagCommand})
	packets.RegisterPacketHandler(17101, []packets.PacketHandler{answer.GetShipDiscuss})
	packets.RegisterPacketHandler(17103, []packets.PacketHandler{answer.PostShipEvaluationComment})
	packets.RegisterPacketHandler(17105, []packets.PacketHandler{answer.ZanShipEvaluation})
	packets.RegisterPacketHandler(17107, []packets.PacketHandler{answer.UpdateShipLike})
	packets.RegisterPacketHandler(17109, []packets.PacketHandler{answer.ReportShipEvaluation})
	packets.RegisterPacketHandler(17301, []packets.PacketHandler{answer.TrophyClaim17301})
	// Dorm / Backyard (190xx)
	packets.RegisterPacketHandler(19002, []packets.PacketHandler{answer.AddDormShip19002})
	packets.RegisterPacketHandler(19004, []packets.PacketHandler{answer.ExitDormShip19004})
	packets.RegisterPacketHandler(19006, []packets.PacketHandler{answer.BuyFurniture19006})
	packets.RegisterPacketHandler(19008, []packets.PacketHandler{answer.PutFurniture19008})
	packets.RegisterPacketHandler(19011, []packets.PacketHandler{answer.ClaimDormIntimacy19011})
	packets.RegisterPacketHandler(19013, []packets.PacketHandler{answer.ClaimDormMoney19013})
	packets.RegisterPacketHandler(19015, []packets.PacketHandler{answer.OpenAddExp19015})
	packets.RegisterPacketHandler(19016, []packets.PacketHandler{answer.RenameDorm19016})
	packets.RegisterPacketHandler(19018, []packets.PacketHandler{answer.GetDormThemeList19018})
	packets.RegisterPacketHandler(19020, []packets.PacketHandler{answer.SaveDormTheme19020})
	packets.RegisterPacketHandler(19022, []packets.PacketHandler{answer.DeleteDormTheme19022})
	packets.RegisterPacketHandler(19024, []packets.PacketHandler{answer.GetBackyardVisitor19024})
	// Backyard theme templates (191xx)
	packets.RegisterPacketHandler(19103, []packets.PacketHandler{answer.GetOSSArgs19103})
	packets.RegisterPacketHandler(19105, []packets.PacketHandler{answer.GetCustomThemeTemplates19105})
	packets.RegisterPacketHandler(19109, []packets.PacketHandler{answer.SaveCustomThemeTemplate19109})
	packets.RegisterPacketHandler(19111, []packets.PacketHandler{answer.PublishCustomThemeTemplate19111})
	packets.RegisterPacketHandler(19113, []packets.PacketHandler{answer.SearchTheme19113})
	packets.RegisterPacketHandler(19115, []packets.PacketHandler{answer.GetCollectionList19115})
	packets.RegisterPacketHandler(19117, []packets.PacketHandler{answer.GetThemeShopList19117})
	packets.RegisterPacketHandler(19119, []packets.PacketHandler{answer.CollectTheme19119})
	packets.RegisterPacketHandler(19121, []packets.PacketHandler{answer.LikeTheme19121})
	packets.RegisterPacketHandler(19123, []packets.PacketHandler{answer.DeleteCustomThemeTemplate19123})
	packets.RegisterPacketHandler(19125, []packets.PacketHandler{answer.UnpublishCustomThemeTemplate19125})
	packets.RegisterPacketHandler(19127, []packets.PacketHandler{answer.CancelCollectTheme19127})
	packets.RegisterPacketHandler(19129, []packets.PacketHandler{answer.InformTheme19129})
	packets.RegisterPacketHandler(19131, []packets.PacketHandler{answer.GetPreviewMd5s19131})
	packets.RegisterPacketHandler(17401, []packets.PacketHandler{answer.ChangeMedalDisplay})
	packets.RegisterPacketHandler(17601, []packets.PacketHandler{answer.EquipCodeShareListRequest})
	packets.RegisterPacketHandler(17603, []packets.PacketHandler{answer.EquipCodeShare})
	packets.RegisterPacketHandler(17605, []packets.PacketHandler{answer.EquipCodeLike})
	packets.RegisterPacketHandler(17607, []packets.PacketHandler{answer.EquipCodeImpeach})
	packets.RegisterPacketHandler(17501, []packets.PacketHandler{answer.UnlockAppreciateGallery})
	packets.RegisterPacketHandler(17503, []packets.PacketHandler{answer.UnlockAppreciateMusic})
	packets.RegisterPacketHandler(17505, []packets.PacketHandler{answer.ToggleAppreciationGalleryLike})
	packets.RegisterPacketHandler(17507, []packets.PacketHandler{answer.ToggleAppreciationMusicLike})
	packets.RegisterPacketHandler(17509, []packets.PacketHandler{answer.MarkMangaRead})
	packets.RegisterPacketHandler(17511, []packets.PacketHandler{answer.ToggleMangaLike})
	packets.RegisterPacketHandler(17513, []packets.PacketHandler{answer.UpdateAppreciationMusicPlayerSettings})
	packets.RegisterPacketHandler(15300, []packets.PacketHandler{func(b *[]byte, c *connection.Client) (int, int, error) {
		return 0, 0, nil
	}})
	packets.RegisterPacketHandler(12299, []packets.PacketHandler{func(b *[]byte, c *connection.Client) (int, int, error) {
		return 0, 0, nil
	}})
}
