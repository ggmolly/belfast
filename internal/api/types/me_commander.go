package types

type MeCommanderResponse struct {
	CommanderID uint32 `json:"commander_id"`
	Name        string `json:"name"`
	Level       uint32 `json:"level"`
}
