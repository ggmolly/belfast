package types

import "time"

type PaginationMeta struct {
	Offset int   `json:"offset"`
	Limit  int   `json:"limit"`
	Total  int64 `json:"total"`
}

type PlayerListResponse struct {
	Players []PlayerSummary `json:"players"`
	Meta    PaginationMeta  `json:"meta"`
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
	FinishesAt string `json:"finishes_at"`
}

type PlayerBuildResponse struct {
	Builds []PlayerBuildEntry `json:"builds"`
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

type GiveSkinRequest struct {
	SkinID uint32 `json:"skin_id" validate:"required,gt=0"`
}

type KickPlayerRequest struct {
	Reason uint8 `json:"reason" validate:"omitempty,oneof=1 2 3 4 5 6 7 199"`
}

type KickPlayerResponse struct {
	Disconnected bool `json:"disconnected"`
}
