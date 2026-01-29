package types

type BuffPayload struct {
	ID          uint32 `json:"id"`
	Name        string `json:"name"`
	Desc        string `json:"desc"`
	MaxTime     int    `json:"max_time"`
	BenefitType string `json:"benefit_type"`
}

type BuffListResponse struct {
	Buffs []BuffPayload  `json:"buffs"`
	Meta  PaginationMeta `json:"meta"`
}
