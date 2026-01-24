package types

import (
	"encoding/json"
	"time"
)

type PaginationMeta struct {
	Offset int   `json:"offset"`
	Limit  int   `json:"limit"`
	Total  int64 `json:"total"`
}

type PlayerListResponse struct {
	Players []PlayerSummary `json:"players"`
	Meta    PaginationMeta  `json:"meta"`
}

type ShipSummary struct {
	ID          uint32  `json:"id"`
	Name        string  `json:"name"`
	RarityID    uint32  `json:"rarity"`
	Star        uint32  `json:"star"`
	Type        uint32  `json:"type"`
	Nationality uint32  `json:"nationality"`
	BuildTime   uint32  `json:"build_time"`
	PoolID      *uint32 `json:"pool_id,omitempty"`
}

type ShipListResponse struct {
	Ships []ShipSummary  `json:"ships"`
	Meta  PaginationMeta `json:"meta"`
}

type ItemSummary struct {
	ID          uint32 `json:"id"`
	Name        string `json:"name"`
	Rarity      int    `json:"rarity"`
	ShopID      int    `json:"shop_id"`
	Type        int    `json:"type"`
	VirtualType int    `json:"virtual_type"`
}

type ItemListResponse struct {
	Items []ItemSummary  `json:"items"`
	Meta  PaginationMeta `json:"meta"`
}

type ResourceSummary struct {
	ID     uint32 `json:"id"`
	ItemID uint32 `json:"item_id"`
	Name   string `json:"name"`
}

type ResourceListResponse struct {
	Resources []ResourceSummary `json:"resources"`
	Meta      PaginationMeta    `json:"meta"`
}

type SkinSummary struct {
	ID        uint32 `json:"id"`
	Name      string `json:"name"`
	ShipGroup int    `json:"ship_group"`
}

type SkinListResponse struct {
	Skins []SkinSummary  `json:"skins"`
	Meta  PaginationMeta `json:"meta"`
}

type ShopOfferSummary struct {
	ID             uint32  `json:"id"`
	Effects        []int64 `json:"effects"`
	Number         int     `json:"num"`
	ResourceNumber int     `json:"resource_num"`
	ResourceID     uint32  `json:"resource_type"`
	Type           uint32  `json:"type"`
}

type ShopOfferListResponse struct {
	Offers []ShopOfferSummary `json:"offers"`
	Meta   PaginationMeta     `json:"meta"`
}

// RawJSON represents arbitrary JSON payloads for swagger generation.
type RawJSON struct {
	Value json.RawMessage `json:"-" swaggerignore:"true"`
}

func (payload RawJSON) MarshalJSON() ([]byte, error) {
	if payload.Value == nil {
		return []byte("null"), nil
	}
	return payload.Value, nil
}

func (payload *RawJSON) UnmarshalJSON(data []byte) error {
	if payload == nil {
		return nil
	}
	payload.Value = json.RawMessage(data)
	return nil
}

type ShoppingStreetState struct {
	Level           uint32 `json:"level"`
	NextFlashTime   uint32 `json:"next_flash_time"`
	LevelUpTime     uint32 `json:"level_up_time"`
	FlashCount      uint32 `json:"flash_count"`
	LastRefreshedAt uint32 `json:"last_refreshed_at"`
}

type ShoppingStreetOfferSummary struct {
	ID             uint32  `json:"id"`
	ResourceNumber int     `json:"resource_num"`
	ResourceID     uint32  `json:"resource_type"`
	Type           uint32  `json:"type"`
	Number         int     `json:"num"`
	Genre          string  `json:"genre"`
	Discount       int     `json:"discount"`
	EffectArgs     RawJSON `json:"effect_args"`
}

type ShoppingStreetGood struct {
	GoodsID  uint32                      `json:"goods_id"`
	Discount uint32                      `json:"discount"`
	BuyCount uint32                      `json:"buy_count"`
	Offer    *ShoppingStreetOfferSummary `json:"offer,omitempty"`
}

type ShoppingStreetResponse struct {
	State ShoppingStreetState  `json:"state"`
	Goods []ShoppingStreetGood `json:"goods"`
}

type ShoppingStreetRefreshRequest struct {
	GoodsCount         *int     `json:"goods_count"`
	NextFlashInSeconds *uint32  `json:"next_flash_in_seconds"`
	SetFlashCount      *uint32  `json:"set_flash_count"`
	Seed               *int64   `json:"seed"`
	GoodsIDs           []uint32 `json:"goods_ids"`
	DiscountOverride   *uint32  `json:"discount_override"`
	BuyCount           *uint32  `json:"buy_count"`
}

type ShoppingStreetUpdateRequest struct {
	Level         *uint32 `json:"level"`
	NextFlashTime *uint32 `json:"next_flash_time"`
	LevelUpTime   *uint32 `json:"level_up_time"`
	FlashCount    *uint32 `json:"flash_count"`
}

type ShoppingStreetGoodInput struct {
	GoodsID  uint32 `json:"goods_id"`
	Discount uint32 `json:"discount"`
	BuyCount uint32 `json:"buy_count"`
}

type ShoppingStreetGoodsReplaceRequest struct {
	Goods []ShoppingStreetGoodInput `json:"goods"`
}

type ShoppingStreetGoodPatchRequest struct {
	Discount *uint32 `json:"discount"`
	BuyCount *uint32 `json:"buy_count"`
}

type ArenaShopState struct {
	FlashCount      uint32 `json:"flash_count"`
	NextFlashTime   uint32 `json:"next_flash_time"`
	LastRefreshTime uint32 `json:"last_refresh_time"`
}

type ArenaShopItem struct {
	ShopID uint32 `json:"shop_id"`
	Count  uint32 `json:"count"`
}

type ArenaShopResponse struct {
	State ArenaShopState  `json:"state"`
	Items []ArenaShopItem `json:"items"`
}

type ArenaShopUpdateRequest struct {
	FlashCount      *uint32 `json:"flash_count"`
	NextFlashTime   *uint32 `json:"next_flash_time"`
	LastRefreshTime *uint32 `json:"last_refresh_time"`
}

type MedalShopState struct {
	NextRefreshTime uint32 `json:"next_refresh_time"`
}

type MedalShopItem struct {
	ID    uint32 `json:"id"`
	Count uint32 `json:"count"`
	Index uint32 `json:"index"`
}

type MedalShopResponse struct {
	State MedalShopState  `json:"state"`
	Items []MedalShopItem `json:"items"`
}

type MedalShopUpdateRequest struct {
	NextRefreshTime *uint32 `json:"next_refresh_time"`
}

type NoticeSummary struct {
	ID         int    `json:"id"`
	Version    string `json:"version"`
	BtnTitle   string `json:"btn_title"`
	Title      string `json:"title"`
	TitleImage string `json:"title_image"`
	TimeDesc   string `json:"time_desc"`
	Content    string `json:"content"`
	TagType    int    `json:"tag_type"`
	Icon       int    `json:"icon"`
	Track      string `json:"track"`
}

type NoticeListResponse struct {
	Notices []NoticeSummary `json:"notices"`
	Meta    PaginationMeta  `json:"meta"`
}

type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type APIResponse[T any] struct {
	OK    bool      `json:"ok"`
	Data  *T        `json:"data,omitempty"`
	Error *APIError `json:"error,omitempty"`
}

type PlayerSummary struct {
	CommanderID uint32 `json:"id"`
	AccountID   uint32 `json:"account_id"`
	Name        string `json:"name"`
	Level       int    `json:"level"`
	LastLogin   string `json:"last_login"`
	Banned      bool   `json:"banned"`
	Online      bool   `json:"online"`
}

type PlayerDetailResponse struct {
	CommanderID uint32 `json:"id"`
	AccountID   uint32 `json:"account_id"`
	Name        string `json:"name"`
	Level       int    `json:"level"`
	Exp         int    `json:"exp"`
	LastLogin   string `json:"last_login"`
	Banned      bool   `json:"banned"`
	Online      bool   `json:"online"`
}

type PlayerResourceEntry struct {
	ResourceID uint32 `json:"resource_id"`
	Amount     uint32 `json:"amount"`
	Name       string `json:"name"`
}

type PlayerResourceResponse struct {
	Resources []PlayerResourceEntry `json:"resources"`
}

type PlayerItemEntry struct {
	ItemID uint32 `json:"item_id"`
	Count  uint32 `json:"count"`
	Name   string `json:"name"`
}

type PlayerItemResponse struct {
	Items []PlayerItemEntry `json:"items"`
}

type PlayerShipEntry struct {
	OwnedID uint32 `json:"owned_id"`
	ShipID  uint32 `json:"ship_id"`
	Level   uint32 `json:"level"`
	Rarity  uint32 `json:"rarity"`
	Name    string `json:"name"`
	SkinID  uint32 `json:"skin_id"`
}

type PlayerShipResponse struct {
	Ships []PlayerShipEntry `json:"ships"`
}

type PlayerBuildEntry struct {
	BuildID    uint32 `json:"build_id"`
	ShipID     uint32 `json:"ship_id"`
	ShipName   string `json:"ship_name"`
	PoolID     uint32 `json:"pool_id"`
	FinishesAt string `json:"finishes_at"`
}

type PlayerBuildResponse struct {
	Builds []PlayerBuildEntry `json:"builds"`
}

type PlayerBuildQueueEntry struct {
	Slot             uint32 `json:"slot"`
	PoolID           uint32 `json:"pool_id"`
	RemainingSeconds uint32 `json:"remaining_seconds"`
	FinishTime       uint32 `json:"finish_time"`
}

type PlayerBuildQueueResponse struct {
	WorklistCount uint32                  `json:"worklist_count"`
	WorklistList  []PlayerBuildQueueEntry `json:"worklist_list"`
	DrawCount1    uint32                  `json:"draw_count_1"`
	DrawCount10   uint32                  `json:"draw_count_10"`
	ExchangeCount uint32                  `json:"exchange_count"`
}

type PlayerBuildCounterUpdateRequest struct {
	DrawCount1    *uint32 `json:"draw_count_1" validate:"omitempty"`
	DrawCount10   *uint32 `json:"draw_count_10" validate:"omitempty"`
	ExchangeCount *uint32 `json:"exchange_count" validate:"omitempty"`
}

type PlayerMailAttachment struct {
	Type     uint32 `json:"type"`
	ItemID   uint32 `json:"item_id"`
	Quantity uint32 `json:"quantity"`
}

type PlayerMailEntry struct {
	MailID      uint32                 `json:"mail_id"`
	Title       string                 `json:"title"`
	Body        string                 `json:"body"`
	Read        bool                   `json:"read"`
	Date        string                 `json:"date"`
	Important   bool                   `json:"important"`
	Archived    bool                   `json:"archived"`
	Sender      *string                `json:"sender,omitempty"`
	Attachments []PlayerMailAttachment `json:"attachments"`
}

type PlayerMailResponse struct {
	Mails []PlayerMailEntry `json:"mails"`
}

type PlayerCompensationAttachment struct {
	Type     uint32 `json:"type"`
	ItemID   uint32 `json:"item_id"`
	Quantity uint32 `json:"quantity"`
}

type PlayerCompensationEntry struct {
	CompensationID uint32                         `json:"compensation_id"`
	Title          string                         `json:"title"`
	Text           string                         `json:"text"`
	SendTime       string                         `json:"send_time"`
	ExpiresAt      string                         `json:"expires_at"`
	AttachFlag     bool                           `json:"attach_flag"`
	Attachments    []PlayerCompensationAttachment `json:"attachments"`
}

type PlayerCompensationResponse struct {
	Compensations []PlayerCompensationEntry `json:"compensations"`
}

type PushCompensationResponse struct {
	Pushed int `json:"pushed"`
	Failed int `json:"failed"`
}

type PlayerFleetEntry struct {
	FleetID uint32   `json:"fleet_id"`
	Name    string   `json:"name"`
	Ships   []uint32 `json:"ships"`
}

type PlayerFleetResponse struct {
	Fleets []PlayerFleetEntry `json:"fleets"`
}

type PlayerSkinEntry struct {
	SkinID    uint32     `json:"skin_id"`
	Name      string     `json:"name"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type PlayerSkinResponse struct {
	Skins []PlayerSkinEntry `json:"skins"`
}

type PlayerBuffEntry struct {
	BuffID    uint32 `json:"buff_id"`
	ExpiresAt string `json:"expires_at"`
}

type PlayerBuffResponse struct {
	Buffs []PlayerBuffEntry `json:"buffs"`
}

type BanPlayerRequest struct {
	Permanent     bool   `json:"permanent"`
	LiftTimestamp string `json:"lift_timestamp" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	DurationSec   *int64 `json:"duration_sec" validate:"omitempty,min=1"`
}

type ResourceUpdateRequest struct {
	Resources []ResourceUpdateEntry `json:"resources" validate:"required,min=1,dive"`
}

type ResourceUpdateEntry struct {
	ResourceID uint32 `json:"resource_id" validate:"required,gt=0"`
	Amount     uint32 `json:"amount" validate:"required"`
}

type GiveShipRequest struct {
	ShipID uint32 `json:"ship_id" validate:"required,gt=0"`
}

type GiveItemRequest struct {
	ItemID uint32 `json:"item_id" validate:"required,gt=0"`
	Amount uint32 `json:"amount" validate:"required,gt=0"`
}

type SendMailRequest struct {
	Title        string                  `json:"title" validate:"required,min=1"`
	Body         string                  `json:"body" validate:"required,min=1"`
	CustomSender *string                 `json:"custom_sender" validate:"omitempty,min=1"`
	Attachments  []SendMailAttachmentDTO `json:"attachments" validate:"omitempty,dive"`
}

type SendMailAttachmentDTO struct {
	Type     uint32 `json:"type" validate:"required,gt=0"`
	ItemID   uint32 `json:"item_id" validate:"required,gt=0"`
	Quantity uint32 `json:"quantity" validate:"required,gt=0"`
}

type CompensationAttachmentDTO struct {
	Type     uint32 `json:"type" validate:"required,gt=0"`
	ItemID   uint32 `json:"item_id" validate:"required,gt=0"`
	Quantity uint32 `json:"quantity" validate:"required,gt=0"`
}

type CreateCompensationRequest struct {
	Title       string                      `json:"title" validate:"required,min=1"`
	Text        string                      `json:"text" validate:"required,min=1"`
	SendTime    string                      `json:"send_time" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	ExpiresAt   string                      `json:"expires_at" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	Attachments []CompensationAttachmentDTO `json:"attachments" validate:"omitempty,dive"`
}

type UpdateCompensationRequest struct {
	Title       *string                      `json:"title" validate:"omitempty,min=1"`
	Text        *string                      `json:"text" validate:"omitempty,min=1"`
	SendTime    *string                      `json:"send_time" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	ExpiresAt   *string                      `json:"expires_at" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	AttachFlag  *bool                        `json:"attach_flag"`
	Attachments *[]CompensationAttachmentDTO `json:"attachments" validate:"omitempty,dive"`
}

type GiveSkinRequest struct {
	SkinID uint32 `json:"skin_id" validate:"required,gt=0"`
}

type PlayerBuffAddRequest struct {
	BuffID    uint32 `json:"buff_id" validate:"required,gt=0"`
	ExpiresAt string `json:"expires_at" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
}

type KickPlayerRequest struct {
	Reason uint8 `json:"reason" validate:"omitempty,oneof=1 2 3 4 5 6 7 199"`
}

type KickPlayerResponse struct {
	Disconnected bool `json:"disconnected"`
}
