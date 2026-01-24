package types

type ConfigEntryPayload struct {
	Key  string  `json:"key"`
	Data RawJSON `json:"data"`
}

type ConfigEntryListResponse struct {
	Entries []ConfigEntryPayload `json:"entries"`
}
