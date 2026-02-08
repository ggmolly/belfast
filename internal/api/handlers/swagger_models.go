package handlers

import "github.com/ggmolly/belfast/internal/api/types"

type OKResponseDoc struct {
	OK bool `json:"ok"`
}

type APIErrorResponseDoc struct {
	OK    bool           `json:"ok"`
	Error types.APIError `json:"error"`
}

type AuthLoginResponseDoc struct {
	OK   bool                    `json:"ok"`
	Data types.AuthLoginResponse `json:"data"`
}

type AuthSessionResponseDoc struct {
	OK   bool                      `json:"ok"`
	Data types.AuthSessionResponse `json:"data"`
}

type AuthBootstrapStatusResponseDoc struct {
	OK   bool                              `json:"ok"`
	Data types.AuthBootstrapStatusResponse `json:"data"`
}

type UserAuthLoginResponseDoc struct {
	OK   bool                        `json:"ok"`
	Data types.UserAuthLoginResponse `json:"data"`
}

type UserAuthSessionResponseDoc struct {
	OK   bool                          `json:"ok"`
	Data types.UserAuthSessionResponse `json:"data"`
}

type UserRegistrationChallengeResponseDoc struct {
	OK   bool                                    `json:"ok"`
	Data types.UserRegistrationChallengeResponse `json:"data"`
}

type UserRegistrationStatusResponseDoc struct {
	OK   bool                                 `json:"ok"`
	Data types.UserRegistrationStatusResponse `json:"data"`
}

type UserPermissionPolicyResponseDoc struct {
	OK   bool                               `json:"ok"`
	Data types.UserPermissionPolicyResponse `json:"data"`
}

type MePermissionsResponseDoc struct {
	OK   bool                        `json:"ok"`
	Data types.MePermissionsResponse `json:"data"`
}

type MeCommanderResponseDoc struct {
	OK   bool                      `json:"ok"`
	Data types.MeCommanderResponse `json:"data"`
}

type AdminUserListResponseDoc struct {
	OK   bool                        `json:"ok"`
	Data types.AdminUserListResponse `json:"data"`
}

type AdminUserResponseDoc struct {
	OK   bool                    `json:"ok"`
	Data types.AdminUserResponse `json:"data"`
}

type RoleListResponseDoc struct {
	OK   bool                   `json:"ok"`
	Data types.RoleListResponse `json:"data"`
}

type PermissionListResponseDoc struct {
	OK   bool                         `json:"ok"`
	Data types.PermissionListResponse `json:"data"`
}

type RolePolicyResponseDoc struct {
	OK   bool                     `json:"ok"`
	Data types.RolePolicyResponse `json:"data"`
}

type AccountRolesResponseDoc struct {
	OK   bool                       `json:"ok"`
	Data types.AccountRolesResponse `json:"data"`
}

type AccountOverridesResponseDoc struct {
	OK   bool                           `json:"ok"`
	Data types.AccountOverridesResponse `json:"data"`
}

type PasskeyListResponseDoc struct {
	OK   bool                      `json:"ok"`
	Data types.PasskeyListResponse `json:"data"`
}

type PasskeyRegisterOptionsResponseDoc struct {
	OK   bool                                 `json:"ok"`
	Data types.PasskeyRegisterOptionsResponse `json:"data"`
}

type PasskeyAuthenticateOptionsResponseDoc struct {
	OK   bool                                     `json:"ok"`
	Data types.PasskeyAuthenticateOptionsResponse `json:"data"`
}

type PasskeyRegisterResponseDoc struct {
	OK   bool                          `json:"ok"`
	Data types.PasskeyRegisterResponse `json:"data"`
}

type ListShipsResponseDoc struct {
	OK   bool                   `json:"ok"`
	Data types.ShipListResponse `json:"data"`
}

type ShipSummaryResponseDoc struct {
	OK   bool              `json:"ok"`
	Data types.ShipSummary `json:"data"`
}

type ShipMutationResponseDoc struct {
	OK bool `json:"ok"`
}

type ListShipTypesResponseDoc struct {
	OK   bool                       `json:"ok"`
	Data types.ShipTypeListResponse `json:"data"`
}

type ShipTypeSummaryResponseDoc struct {
	OK   bool                  `json:"ok"`
	Data types.ShipTypeSummary `json:"data"`
}

type ShipTypeMutationResponseDoc struct {
	OK bool `json:"ok"`
}

type ListRaritiesResponseDoc struct {
	OK   bool                     `json:"ok"`
	Data types.RarityListResponse `json:"data"`
}

type RaritySummaryResponseDoc struct {
	OK   bool                `json:"ok"`
	Data types.RaritySummary `json:"data"`
}

type RarityMutationResponseDoc struct {
	OK bool `json:"ok"`
}

type ListSkinsResponseDoc struct {
	OK   bool                   `json:"ok"`
	Data types.SkinListResponse `json:"data"`
}

type SkinSummaryResponseDoc struct {
	OK   bool              `json:"ok"`
	Data types.SkinSummary `json:"data"`
}

type SkinDetailResponseDoc struct {
	OK   bool              `json:"ok"`
	Data types.SkinPayload `json:"data"`
}

type SkinMutationResponseDoc struct {
	OK bool `json:"ok"`
}

type ListSkinRestrictionsResponseDoc struct {
	OK   bool                              `json:"ok"`
	Data types.SkinRestrictionListResponse `json:"data"`
}

type SkinRestrictionResponseDoc struct {
	OK   bool                         `json:"ok"`
	Data types.SkinRestrictionPayload `json:"data"`
}

type SkinRestrictionMutationResponseDoc struct {
	OK bool `json:"ok"`
}

type ListSkinRestrictionWindowsResponseDoc struct {
	OK   bool                                    `json:"ok"`
	Data types.SkinRestrictionWindowListResponse `json:"data"`
}

type SkinRestrictionWindowResponseDoc struct {
	OK   bool                               `json:"ok"`
	Data types.SkinRestrictionWindowPayload `json:"data"`
}

type SkinRestrictionWindowMutationResponseDoc struct {
	OK bool `json:"ok"`
}

type ListItemsResponseDoc struct {
	OK   bool                   `json:"ok"`
	Data types.ItemListResponse `json:"data"`
}

type ItemSummaryResponseDoc struct {
	OK   bool              `json:"ok"`
	Data types.ItemSummary `json:"data"`
}

type ItemMutationResponseDoc struct {
	OK bool `json:"ok"`
}

type ListResourcesResponseDoc struct {
	OK   bool                       `json:"ok"`
	Data types.ResourceListResponse `json:"data"`
}

type ResourceSummaryResponseDoc struct {
	OK   bool                  `json:"ok"`
	Data types.ResourceSummary `json:"data"`
}

type ResourceMutationResponseDoc struct {
	OK bool `json:"ok"`
}

type RequisitionShipListResponseDoc struct {
	OK   bool                              `json:"ok"`
	Data types.RequisitionShipListResponse `json:"data"`
}

type RequisitionShipMutationResponseDoc struct {
	OK bool `json:"ok"`
}

type ListEquipmentResponseDoc struct {
	OK   bool                        `json:"ok"`
	Data types.EquipmentListResponse `json:"data"`
}

type EquipmentDetailResponseDoc struct {
	OK   bool                   `json:"ok"`
	Data types.EquipmentPayload `json:"data"`
}

type EquipmentMutationResponseDoc struct {
	OK bool `json:"ok"`
}

type ListWeaponsResponseDoc struct {
	OK   bool                     `json:"ok"`
	Data types.WeaponListResponse `json:"data"`
}

type WeaponDetailResponseDoc struct {
	OK   bool                `json:"ok"`
	Data types.WeaponPayload `json:"data"`
}

type WeaponMutationResponseDoc struct {
	OK bool `json:"ok"`
}

type ListSkillsResponseDoc struct {
	OK   bool                    `json:"ok"`
	Data types.SkillListResponse `json:"data"`
}

type SkillDetailResponseDoc struct {
	OK   bool               `json:"ok"`
	Data types.SkillPayload `json:"data"`
}

type SkillMutationResponseDoc struct {
	OK bool `json:"ok"`
}

type ListBuffsResponseDoc struct {
	OK   bool                   `json:"ok"`
	Data types.BuffListResponse `json:"data"`
}

type BuffDetailResponseDoc struct {
	OK   bool              `json:"ok"`
	Data types.BuffPayload `json:"data"`
}

type BuffMutationResponseDoc struct {
	OK bool `json:"ok"`
}

type ListPlayersResponseDoc struct {
	OK   bool                     `json:"ok"`
	Data types.PlayerListResponse `json:"data"`
}

type PlayerDetailResponseDoc struct {
	OK   bool                       `json:"ok"`
	Data types.PlayerDetailResponse `json:"data"`
}

type PlayerMutationResponseDoc struct {
	OK   bool                         `json:"ok"`
	Data types.PlayerMutationResponse `json:"data"`
}

type PlayerResourcesResponseDoc struct {
	OK   bool                         `json:"ok"`
	Data types.PlayerResourceResponse `json:"data"`
}

type PlayerResourceEntryResponseDoc struct {
	OK   bool                      `json:"ok"`
	Data types.PlayerResourceEntry `json:"data"`
}

type PlayerItemsResponseDoc struct {
	OK   bool                     `json:"ok"`
	Data types.PlayerItemResponse `json:"data"`
}

type PlayerItemEntryResponseDoc struct {
	OK   bool                  `json:"ok"`
	Data types.PlayerItemEntry `json:"data"`
}

type PlayerEquipmentResponseDoc struct {
	OK   bool                          `json:"ok"`
	Data types.PlayerEquipmentResponse `json:"data"`
}

type PlayerEquipmentEntryResponseDoc struct {
	OK   bool                       `json:"ok"`
	Data types.PlayerEquipmentEntry `json:"data"`
}

type PlayerShipEquipmentResponseDoc struct {
	OK   bool                              `json:"ok"`
	Data types.PlayerShipEquipmentResponse `json:"data"`
}

type PlayerMiscItemsResponseDoc struct {
	OK   bool                         `json:"ok"`
	Data types.PlayerMiscItemResponse `json:"data"`
}

type PlayerMiscItemEntryResponseDoc struct {
	OK   bool                      `json:"ok"`
	Data types.PlayerMiscItemEntry `json:"data"`
}

type PlayerRemasterStateResponseDoc struct {
	OK   bool                              `json:"ok"`
	Data types.PlayerRemasterStateResponse `json:"data"`
}

type PlayerRemasterProgressResponseDoc struct {
	OK   bool                                 `json:"ok"`
	Data types.PlayerRemasterProgressResponse `json:"data"`
}

type PlayerChapterStateResponseDoc struct {
	OK   bool                             `json:"ok"`
	Data types.PlayerChapterStateResponse `json:"data"`
}

type PlayerChapterStateListResponseDoc struct {
	OK   bool                                 `json:"ok"`
	Data types.PlayerChapterStateListResponse `json:"data"`
}

type PlayerChapterProgressResponseDoc struct {
	OK   bool                                `json:"ok"`
	Data types.PlayerChapterProgressResponse `json:"data"`
}

type PlayerChapterProgressListResponseDoc struct {
	OK   bool                                    `json:"ok"`
	Data types.PlayerChapterProgressListResponse `json:"data"`
}

type PlayerShipsResponseDoc struct {
	OK   bool                     `json:"ok"`
	Data types.PlayerShipResponse `json:"data"`
}

type PlayerSecretariesResponseDoc struct {
	OK   bool                            `json:"ok"`
	Data types.PlayerSecretariesResponse `json:"data"`
}

type PlayerOwnedShipEntryResponseDoc struct {
	OK   bool                       `json:"ok"`
	Data types.PlayerOwnedShipEntry `json:"data"`
}

type PlayerBuildsResponseDoc struct {
	OK   bool                      `json:"ok"`
	Data types.PlayerBuildResponse `json:"data"`
}

type PlayerBuildEntryResponseDoc struct {
	OK   bool                   `json:"ok"`
	Data types.PlayerBuildEntry `json:"data"`
}

type PlayerBuildQueueResponseDoc struct {
	OK   bool                           `json:"ok"`
	Data types.PlayerBuildQueueResponse `json:"data"`
}

type PlayerSupportRequisitionResponseDoc struct {
	OK   bool                                   `json:"ok"`
	Data types.PlayerSupportRequisitionResponse `json:"data"`
}

type PlayerMailsResponseDoc struct {
	OK   bool                     `json:"ok"`
	Data types.PlayerMailResponse `json:"data"`
}

type PlayerMailEntryResponseDoc struct {
	OK   bool                  `json:"ok"`
	Data types.PlayerMailEntry `json:"data"`
}

type PlayerCompensationsResponseDoc struct {
	OK   bool                             `json:"ok"`
	Data types.PlayerCompensationResponse `json:"data"`
}

type PlayerPunishmentsResponseDoc struct {
	OK   bool                            `json:"ok"`
	Data types.PlayerPunishmentsResponse `json:"data"`
}

type PlayerPunishmentEntryResponseDoc struct {
	OK   bool                        `json:"ok"`
	Data types.PlayerPunishmentEntry `json:"data"`
}

type PushCompensationResponseDoc struct {
	OK   bool                           `json:"ok"`
	Data types.PushCompensationResponse `json:"data"`
}

type PlayerFleetsResponseDoc struct {
	OK   bool                      `json:"ok"`
	Data types.PlayerFleetResponse `json:"data"`
}

type PlayerFleetEntryResponseDoc struct {
	OK   bool                   `json:"ok"`
	Data types.PlayerFleetEntry `json:"data"`
}

type PlayerSkinsResponseDoc struct {
	OK   bool                     `json:"ok"`
	Data types.PlayerSkinResponse `json:"data"`
}

type PlayerSkinEntryResponseDoc struct {
	OK   bool                  `json:"ok"`
	Data types.PlayerSkinEntry `json:"data"`
}

type PlayerBuffsResponseDoc struct {
	OK   bool                     `json:"ok"`
	Data types.PlayerBuffResponse `json:"data"`
}

type PlayerBuffEntryResponseDoc struct {
	OK   bool                  `json:"ok"`
	Data types.PlayerBuffEntry `json:"data"`
}

type PlayerFlagsResponseDoc struct {
	OK   bool                      `json:"ok"`
	Data types.PlayerFlagsResponse `json:"data"`
}

type PlayerRandomFlagShipModeResponseDoc struct {
	OK   bool                                   `json:"ok"`
	Data types.PlayerRandomFlagShipModeResponse `json:"data"`
}

type PlayerRandomFlagShipResponseDoc struct {
	OK   bool                               `json:"ok"`
	Data types.PlayerRandomFlagShipResponse `json:"data"`
}

type PlayerRandomFlagShipListResponseDoc struct {
	OK   bool                                   `json:"ok"`
	Data types.PlayerRandomFlagShipListResponse `json:"data"`
}

type PlayerGuideResponseDoc struct {
	OK   bool                      `json:"ok"`
	Data types.PlayerGuideResponse `json:"data"`
}

type PlayerStoriesResponseDoc struct {
	OK   bool                        `json:"ok"`
	Data types.PlayerStoriesResponse `json:"data"`
}

type PlayerLikesResponseDoc struct {
	OK   bool                      `json:"ok"`
	Data types.PlayerLikesResponse `json:"data"`
}

type PlayerAttiresResponseDoc struct {
	OK   bool                       `json:"ok"`
	Data types.PlayerAttireResponse `json:"data"`
}

type PlayerLivingAreaCoverResponseDoc struct {
	OK   bool                                `json:"ok"`
	Data types.PlayerLivingAreaCoverResponse `json:"data"`
}

type CommanderTBResponseDoc struct {
	OK   bool                     `json:"ok"`
	Data types.CommanderTBPayload `json:"data"`
}

type ConfigEntryListResponseDoc struct {
	OK   bool                          `json:"ok"`
	Data types.ConfigEntryListResponse `json:"data"`
}

type ConfigEntryResponseDoc struct {
	OK   bool                     `json:"ok"`
	Data types.ConfigEntryPayload `json:"data"`
}

type ConfigEntryMutationResponseDoc struct {
	OK bool `json:"ok"`
}

type ActivityAllowlistResponseDoc struct {
	OK   bool                           `json:"ok"`
	Data types.ActivityAllowlistPayload `json:"data"`
}

type PlayerShoppingStreetResponseDoc struct {
	OK   bool                         `json:"ok"`
	Data types.ShoppingStreetResponse `json:"data"`
}

type PlayerShoppingStreetGoodsResponseDoc struct {
	OK   bool                              `json:"ok"`
	Data types.ShoppingStreetGoodsResponse `json:"data"`
}

type PlayerShoppingStreetGoodResponseDoc struct {
	OK   bool                     `json:"ok"`
	Data types.ShoppingStreetGood `json:"data"`
}

type PlayerArenaShopResponseDoc struct {
	OK   bool                    `json:"ok"`
	Data types.ArenaShopResponse `json:"data"`
}

type PlayerArenaShopDeleteResponseDoc struct {
	OK bool `json:"ok"`
}

type PlayerMedalShopResponseDoc struct {
	OK   bool                    `json:"ok"`
	Data types.MedalShopResponse `json:"data"`
}

type PlayerMedalShopGoodsResponseDoc struct {
	OK   bool                         `json:"ok"`
	Data types.MedalShopGoodsResponse `json:"data"`
}

type PlayerGuildShopResponseDoc struct {
	OK   bool                    `json:"ok"`
	Data types.GuildShopResponse `json:"data"`
}

type PlayerGuildShopGoodsResponseDoc struct {
	OK   bool                         `json:"ok"`
	Data types.GuildShopGoodsResponse `json:"data"`
}

type PlayerMiniGameShopResponseDoc struct {
	OK   bool                       `json:"ok"`
	Data types.MiniGameShopResponse `json:"data"`
}

type PlayerMiniGameShopGoodsResponseDoc struct {
	OK   bool                            `json:"ok"`
	Data types.MiniGameShopGoodsResponse `json:"data"`
}

type ShopOfferListResponseDoc struct {
	OK   bool                        `json:"ok"`
	Data types.ShopOfferListResponse `json:"data"`
}

type NoticeListResponseDoc struct {
	OK   bool                     `json:"ok"`
	Data types.NoticeListResponse `json:"data"`
}

type NoticeSummaryResponseDoc struct {
	OK   bool                `json:"ok"`
	Data types.NoticeSummary `json:"data"`
}

type NoticeActiveResponseDoc struct {
	OK   bool                  `json:"ok"`
	Data []types.NoticeSummary `json:"data"`
}

type ExchangeCodeListResponseDoc struct {
	OK   bool                           `json:"ok"`
	Data types.ExchangeCodeListResponse `json:"data"`
}

type ExchangeCodeSummaryResponseDoc struct {
	OK   bool                      `json:"ok"`
	Data types.ExchangeCodeSummary `json:"data"`
}

type ExchangeCodeRedeemListResponseDoc struct {
	OK   bool                                 `json:"ok"`
	Data types.ExchangeCodeRedeemListResponse `json:"data"`
}

type ServerStatusResponseDoc struct {
	OK   bool                       `json:"ok"`
	Data types.ServerStatusResponse `json:"data"`
}

type ServerConfigResponseDoc struct {
	OK   bool                       `json:"ok"`
	Data types.ServerConfigResponse `json:"data"`
}

type ServerMaintenanceResponseDoc struct {
	OK   bool                            `json:"ok"`
	Data types.ServerMaintenanceResponse `json:"data"`
}

type ServerStatsResponseDoc struct {
	OK   bool                      `json:"ok"`
	Data types.ServerStatsResponse `json:"data"`
}

type ServerMetricsResponseDoc struct {
	OK   bool                        `json:"ok"`
	Data types.ServerMetricsResponse `json:"data"`
}

type ServerUptimeResponseDoc struct {
	OK   bool                       `json:"ok"`
	Data types.ServerUptimeResponse `json:"data"`
}

type ConnectionListResponseDoc struct {
	OK   bool                      `json:"ok"`
	Data []types.ConnectionSummary `json:"data"`
}

type ConnectionDetailResponseDoc struct {
	OK   bool                   `json:"ok"`
	Data types.ConnectionDetail `json:"data"`
}

type Dorm3dApartmentListResponseDoc struct {
	OK   bool                              `json:"ok"`
	Data types.Dorm3dApartmentListResponse `json:"data"`
}

type Dorm3dApartmentResponseDoc struct {
	OK   bool                  `json:"ok"`
	Data types.Dorm3dApartment `json:"data"`
}

type Dorm3dApartmentGiftsResponseDoc struct {
	OK   bool                 `json:"ok"`
	Data types.Dorm3dGiftList `json:"data"`
}

type Dorm3dApartmentShipsResponseDoc struct {
	OK   bool                 `json:"ok"`
	Data types.Dorm3dShipList `json:"data"`
}

type Dorm3dApartmentRoomsResponseDoc struct {
	OK   bool                 `json:"ok"`
	Data types.Dorm3dRoomList `json:"data"`
}

type Dorm3dApartmentInsResponseDoc struct {
	OK   bool                `json:"ok"`
	Data types.Dorm3dInsList `json:"data"`
}

type JuustagramTemplateListResponseDoc struct {
	OK   bool                                 `json:"ok"`
	Data types.JuustagramTemplateListResponse `json:"data"`
}

type JuustagramTemplateResponseDoc struct {
	OK   bool                     `json:"ok"`
	Data types.JuustagramTemplate `json:"data"`
}

type JuustagramNpcTemplateListResponseDoc struct {
	OK   bool                                    `json:"ok"`
	Data types.JuustagramNpcTemplateListResponse `json:"data"`
}

type JuustagramNpcTemplateResponseDoc struct {
	OK   bool                        `json:"ok"`
	Data types.JuustagramNpcTemplate `json:"data"`
}

type JuustagramShipGroupListResponseDoc struct {
	OK   bool                                  `json:"ok"`
	Data types.JuustagramShipGroupListResponse `json:"data"`
}

type JuustagramShipGroupResponseDoc struct {
	OK   bool                              `json:"ok"`
	Data types.JuustagramShipGroupTemplate `json:"data"`
}

type JuustagramLanguageResponseDoc struct {
	OK   bool                     `json:"ok"`
	Data types.JuustagramLanguage `json:"data"`
}

type JuustagramLanguageListResponseDoc struct {
	OK   bool                                 `json:"ok"`
	Data types.JuustagramLanguageListResponse `json:"data"`
}

type JuustagramMessageListResponseDoc struct {
	OK   bool                                `json:"ok"`
	Data types.JuustagramMessageListResponse `json:"data"`
}

type JuustagramMessageResponseDoc struct {
	OK   bool                            `json:"ok"`
	Data types.JuustagramMessageResponse `json:"data"`
}

type JuustagramMessageStateListResponseDoc struct {
	OK   bool                                     `json:"ok"`
	Data types.JuustagramMessageStateListResponse `json:"data"`
}

type JuustagramMessageStateResponseDoc struct {
	OK   bool                                 `json:"ok"`
	Data types.JuustagramMessageStateResponse `json:"data"`
}

type JuustagramDiscussResponseDoc struct {
	OK   bool                            `json:"ok"`
	Data types.JuustagramDiscussResponse `json:"data"`
}

type JuustagramPlayerDiscussListResponseDoc struct {
	OK   bool                                      `json:"ok"`
	Data types.JuustagramPlayerDiscussListResponse `json:"data"`
}

type JuustagramPlayerDiscussResponseDoc struct {
	OK   bool                                  `json:"ok"`
	Data types.JuustagramPlayerDiscussResponse `json:"data"`
}

type JuustagramGroupListResponseDoc struct {
	OK   bool                              `json:"ok"`
	Data types.JuustagramGroupListResponse `json:"data"`
}

type JuustagramGroupResponseDoc struct {
	OK   bool                          `json:"ok"`
	Data types.JuustagramGroupResponse `json:"data"`
}

type JuustagramGroupDeleteResponseDoc struct {
	OK bool `json:"ok"`
}

type JuustagramChatGroupDeleteResponseDoc struct {
	OK bool `json:"ok"`
}

type JuustagramChatReplyDeleteResponseDoc struct {
	OK bool `json:"ok"`
}

type KickPlayerResponseDoc struct {
	OK   bool                     `json:"ok"`
	Data types.KickPlayerResponse `json:"data"`
}
