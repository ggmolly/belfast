package types

type ConfigEntryPayload struct {
	ID       uint64  `json:"id"`
	Category string  `json:"category"`
	Key      string  `json:"key"`
	Data     RawJSON `json:"data"`
}

type ConfigEntryMutationRequest struct {
	Category string  `json:"category"`
	Key      string  `json:"key"`
	Data     RawJSON `json:"data"`
}

type ConfigEntryListResponse struct {
	Entries []ConfigEntryPayload `json:"entries"`
}
