package types

type CommanderTBPayload struct {
	CommanderID uint32  `json:"commander_id"`
	Tb          RawJSON `json:"tb"`
	Permanent   RawJSON `json:"permanent"`
}

type CommanderTBRequest struct {
	Tb        RawJSON `json:"tb"`
	Permanent RawJSON `json:"permanent"`
}
