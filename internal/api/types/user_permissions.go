package types

type UserPermissionPolicyResponse struct {
	Actions          []string `json:"actions"`
	AvailableActions []string `json:"available_actions"`
	UpdatedAt        string   `json:"updated_at"`
	UpdatedBy        string   `json:"updated_by"`
}

type UserPermissionPolicyUpdateRequest struct {
	Actions []string `json:"actions"`
}
