package types

type ExchangeReward struct {
	Type  uint32 `json:"type"`
	ID    uint32 `json:"id"`
	Count uint32 `json:"count"`
}

type ExchangeCodeSummary struct {
	ID       uint32           `json:"id"`
	Code     string           `json:"code"`
	Platform string           `json:"platform"`
	Quota    int              `json:"quota"`
	Rewards  []ExchangeReward `json:"rewards"`
}

type ExchangeCodeListResponse struct {
	Codes []ExchangeCodeSummary `json:"codes"`
	Meta  PaginationMeta        `json:"meta"`
}

type ExchangeCodeRequest struct {
	Code     string           `json:"code"`
	Platform string           `json:"platform"`
	Quota    *int             `json:"quota"`
	Rewards  []ExchangeReward `json:"rewards"`
}
