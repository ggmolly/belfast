package types

type RequisitionShipRequest struct {
	ShipID uint32 `json:"ship_id"`
}

type RequisitionShipListResponse struct {
	ShipIDs []uint32 `json:"ship_ids"`
}
