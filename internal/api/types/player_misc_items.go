package types

type PlayerMiscItemEntry struct {
	ItemID uint32 `json:"item_id"`
	Data   uint32 `json:"data"`
	Name   string `json:"name"`
}

type PlayerMiscItemResponse struct {
	Items []PlayerMiscItemEntry `json:"items"`
}

type PlayerMiscItemUpdateRequest struct {
	Data *uint32 `json:"data" validate:"required"`
}
