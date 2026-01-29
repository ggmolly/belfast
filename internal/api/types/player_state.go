package types

import "time"

type PlayerFlagsResponse struct {
	Flags []uint32 `json:"flags"`
}

type PlayerFlagRequest struct {
	FlagID uint32 `json:"flag_id" validate:"required,gt=0"`
}

type PlayerRandomFlagShipModeResponse struct {
	Mode uint32 `json:"mode"`
}

type PlayerRandomFlagShipModeRequest struct {
	Mode uint32 `json:"mode" validate:"required,oneof=1 2 3"`
}

type PlayerRandomFlagShipResponse struct {
	Enabled bool `json:"enabled"`
}

type PlayerRandomFlagShipRequest struct {
	Enabled bool `json:"enabled"`
}

type PlayerRandomFlagShipEntry struct {
	ShipID    uint32 `json:"ship_id"`
	PhantomID uint32 `json:"phantom_id"`
	Enabled   bool   `json:"enabled"`
}

type PlayerRandomFlagShipListResponse struct {
	Entries []PlayerRandomFlagShipEntry `json:"entries"`
}

type PlayerRandomFlagShipUpsertRequest struct {
	ShipID    uint32 `json:"ship_id" validate:"required,gt=0"`
	PhantomID uint32 `json:"phantom_id" validate:"required,gt=0"`
	Enabled   bool   `json:"enabled"`
}

type PlayerSecretaryEntry struct {
	ShipID      uint32  `json:"ship_id"`
	PhantomID   uint32  `json:"phantom_id"`
	IsSecretary bool    `json:"is_secretary"`
	Position    *uint32 `json:"position,omitempty"`
}

type PlayerSecretariesResponse struct {
	Ships []PlayerSecretaryEntry `json:"ships"`
}

type PlayerSecretaryUpdate struct {
	ShipID    uint32 `json:"ship_id" validate:"required,gt=0"`
	PhantomID uint32 `json:"phantom_id"`
}

type PlayerSecretariesReplaceRequest struct {
	Secretaries []PlayerSecretaryUpdate `json:"secretaries" validate:"required,dive"`
}

type PlayerGuideResponse struct {
	GuideIndex    uint32 `json:"guide_index"`
	NewGuideIndex uint32 `json:"new_guide_index"`
}

type PlayerGuideUpdateRequest struct {
	GuideIndex    *uint32 `json:"guide_index" validate:"omitempty"`
	NewGuideIndex *uint32 `json:"new_guide_index" validate:"omitempty"`
}

type PlayerStoriesResponse struct {
	Stories []uint32 `json:"stories"`
}

type PlayerStoryRequest struct {
	StoryID uint32 `json:"story_id" validate:"required,gt=0"`
}

type PlayerLikesResponse struct {
	GroupIDs []uint32 `json:"group_ids"`
}

type PlayerLikeRequest struct {
	GroupID uint32 `json:"group_id" validate:"required,gt=0"`
}

type PlayerAttireEntry struct {
	Type      uint32     `json:"type"`
	AttireID  uint32     `json:"attire_id"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	IsNew     bool       `json:"is_new"`
}

type PlayerAttireResponse struct {
	Attires []PlayerAttireEntry `json:"attires"`
}

type PlayerAttireCreateRequest struct {
	Type      uint32  `json:"type" validate:"required,gt=0"`
	AttireID  uint32  `json:"attire_id" validate:"required,gt=0"`
	ExpiresAt *string `json:"expires_at" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	IsNew     *bool   `json:"is_new"`
}

type PlayerAttireUpdateRequest struct {
	ExpiresAt *string `json:"expires_at" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	IsNew     *bool   `json:"is_new"`
}

type PlayerAttireSelectionUpdateRequest struct {
	IconFrameID *uint32 `json:"icon_frame_id" validate:"omitempty"`
	ChatFrameID *uint32 `json:"chat_frame_id" validate:"omitempty"`
	BattleUIID  *uint32 `json:"battle_ui_id" validate:"omitempty"`
}

type PlayerLivingAreaCoverResponse struct {
	Selected uint32   `json:"selected"`
	Owned    []uint32 `json:"owned"`
}

type PlayerLivingAreaCoverRequest struct {
	CoverID uint32 `json:"cover_id" validate:"required,gt=0"`
}

type PlayerLivingAreaCoverSelectRequest struct {
	CoverID uint32 `json:"cover_id" validate:"required"`
}

type PlayerLivingAreaCoverUpdateRequest struct {
	IsNew *bool `json:"is_new"`
}
