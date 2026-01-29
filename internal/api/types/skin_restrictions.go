package types

type SkinRestrictionPayload struct {
	SkinID uint32 `json:"skin_id"`
	Type   uint32 `json:"type"`
}

type SkinRestrictionCreateRequest struct {
	SkinID uint32 `json:"skin_id"`
	Type   uint32 `json:"type"`
}

type SkinRestrictionUpdateRequest struct {
	Type uint32 `json:"type"`
}

type SkinRestrictionListResponse struct {
	SkinRestrictions []SkinRestrictionPayload `json:"skin_restrictions"`
	Meta             PaginationMeta           `json:"meta"`
}

type SkinRestrictionWindowPayload struct {
	ID        uint32 `json:"id"`
	SkinID    uint32 `json:"skin_id"`
	Type      uint32 `json:"type"`
	StartTime uint32 `json:"start_time"`
	StopTime  uint32 `json:"stop_time"`
}

type SkinRestrictionWindowCreateRequest struct {
	ID        uint32 `json:"id"`
	SkinID    uint32 `json:"skin_id"`
	Type      uint32 `json:"type"`
	StartTime uint32 `json:"start_time"`
	StopTime  uint32 `json:"stop_time"`
}

type SkinRestrictionWindowUpdateRequest struct {
	SkinID    uint32 `json:"skin_id"`
	Type      uint32 `json:"type"`
	StartTime uint32 `json:"start_time"`
	StopTime  uint32 `json:"stop_time"`
}

type SkinRestrictionWindowListResponse struct {
	Windows []SkinRestrictionWindowPayload `json:"windows"`
	Meta    PaginationMeta                 `json:"meta"`
}
