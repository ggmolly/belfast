package types

type PermissionPolicyEntry struct {
	Key       string `json:"key"`
	ReadSelf  bool   `json:"read_self"`
	ReadAny   bool   `json:"read_any"`
	WriteSelf bool   `json:"write_self"`
	WriteAny  bool   `json:"write_any"`
}

type UserPermissionPolicyResponse struct {
	Role          string                  `json:"role"`
	Permissions   []PermissionPolicyEntry `json:"permissions"`
	AvailableKeys []string                `json:"available_keys"`
	UpdatedAt     string                  `json:"updated_at"`
	UpdatedBy     string                  `json:"updated_by"`
}

type UserPermissionPolicyUpdateRequest struct {
	Permissions []PermissionPolicyEntry `json:"permissions"`
}
