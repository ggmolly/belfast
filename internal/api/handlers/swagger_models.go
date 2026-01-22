package handlers

import "github.com/ggmolly/belfast/internal/api/types"

type OKResponseDoc struct {
	OK bool `json:"ok"`
}

type APIErrorResponseDoc struct {
	OK    bool           `json:"ok"`
	Error types.APIError `json:"error"`
}

type ListShipsResponseDoc struct {
	OK   bool                   `json:"ok"`
	Data types.ShipListResponse `json:"data"`
}

type ShipSummaryResponseDoc struct {
	OK   bool              `json:"ok"`
	Data types.ShipSummary `json:"data"`
}

type ListSkinsResponseDoc struct {
	OK   bool                   `json:"ok"`
	Data types.SkinListResponse `json:"data"`
}

type SkinSummaryResponseDoc struct {
	OK   bool              `json:"ok"`
	Data types.SkinSummary `json:"data"`
}

type ListItemsResponseDoc struct {
	OK   bool                   `json:"ok"`
	Data types.ItemListResponse `json:"data"`
}

type ItemSummaryResponseDoc struct {
	OK   bool              `json:"ok"`
	Data types.ItemSummary `json:"data"`
}

type ListResourcesResponseDoc struct {
	OK   bool                       `json:"ok"`
	Data types.ResourceListResponse `json:"data"`
}

type ResourceSummaryResponseDoc struct {
	OK   bool                  `json:"ok"`
	Data types.ResourceSummary `json:"data"`
}

type ListPlayersResponseDoc struct {
	OK   bool                     `json:"ok"`
	Data types.PlayerListResponse `json:"data"`
}

type PlayerDetailResponseDoc struct {
	OK   bool                       `json:"ok"`
	Data types.PlayerDetailResponse `json:"data"`
}

type PlayerResourcesResponseDoc struct {
	OK   bool                         `json:"ok"`
	Data types.PlayerResourceResponse `json:"data"`
}

type PlayerItemsResponseDoc struct {
	OK   bool                     `json:"ok"`
	Data types.PlayerItemResponse `json:"data"`
}

type PlayerShipsResponseDoc struct {
	OK   bool                     `json:"ok"`
	Data types.PlayerShipResponse `json:"data"`
}

type PlayerBuildsResponseDoc struct {
	OK   bool                      `json:"ok"`
	Data types.PlayerBuildResponse `json:"data"`
}

type PlayerMailsResponseDoc struct {
	OK   bool                     `json:"ok"`
	Data types.PlayerMailResponse `json:"data"`
}

type PlayerCompensationsResponseDoc struct {
	OK   bool                             `json:"ok"`
	Data types.PlayerCompensationResponse `json:"data"`
}

type PushCompensationResponseDoc struct {
	OK   bool                           `json:"ok"`
	Data types.PushCompensationResponse `json:"data"`
}

type PlayerFleetsResponseDoc struct {
	OK   bool                      `json:"ok"`
	Data types.PlayerFleetResponse `json:"data"`
}

type PlayerSkinsResponseDoc struct {
	OK   bool                     `json:"ok"`
	Data types.PlayerSkinResponse `json:"data"`
}

type PlayerBuffsResponseDoc struct {
	OK   bool                     `json:"ok"`
	Data types.PlayerBuffResponse `json:"data"`
}

type PlayerShoppingStreetResponseDoc struct {
	OK   bool                         `json:"ok"`
	Data types.ShoppingStreetResponse `json:"data"`
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

type ServerMaintenanceResponseDoc struct {
	OK   bool                            `json:"ok"`
	Data types.ServerMaintenanceResponse `json:"data"`
}

type Dorm3dApartmentListResponseDoc struct {
	OK   bool                              `json:"ok"`
	Data types.Dorm3dApartmentListResponse `json:"data"`
}

type Dorm3dApartmentResponseDoc struct {
	OK   bool                  `json:"ok"`
	Data types.Dorm3dApartment `json:"data"`
}
