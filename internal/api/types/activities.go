package types

type ActivityAllowlistPayload struct {
	IDs []uint32 `json:"ids"`
}

type ActivityAllowlistPatchPayload struct {
	Add    []uint32 `json:"add"`
	Remove []uint32 `json:"remove"`
}
