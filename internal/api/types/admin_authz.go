package types

type RoleSummary struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UpdatedAt   string `json:"updated_at"`
	UpdatedBy   string `json:"updated_by"`
}

type RoleListResponse struct {
	Roles []RoleSummary `json:"roles"`
}

type PermissionSummary struct {
	Key         string `json:"key"`
	Description string `json:"description"`
}

type PermissionListResponse struct {
	Permissions []PermissionSummary `json:"permissions"`
}

type RolePolicyResponse struct {
	Role          string                  `json:"role"`
	Permissions   []PermissionPolicyEntry `json:"permissions"`
	AvailableKeys []string                `json:"available_keys"`
	UpdatedAt     string                  `json:"updated_at"`
	UpdatedBy     string                  `json:"updated_by"`
}

type RolePolicyUpdateRequest struct {
	Permissions []PermissionPolicyEntry `json:"permissions"`
}

type AccountRolesResponse struct {
	AccountID string   `json:"account_id"`
	Roles     []string `json:"roles"`
}

type AccountRolesUpdateRequest struct {
	Roles []string `json:"roles"`
}

type AccountOverrideEntry struct {
	Key       string `json:"key"`
	Mode      string `json:"mode"` // allow|deny
	ReadSelf  bool   `json:"read_self"`
	ReadAny   bool   `json:"read_any"`
	WriteSelf bool   `json:"write_self"`
	WriteAny  bool   `json:"write_any"`
}

type AccountOverridesResponse struct {
	AccountID string                 `json:"account_id"`
	Overrides []AccountOverrideEntry `json:"overrides"`
}

type AccountOverridesUpdateRequest struct {
	Overrides []AccountOverrideEntry `json:"overrides"`
}
