package types

type MePermissionsResponse struct {
	Roles       []string                `json:"roles"`
	Permissions []PermissionPolicyEntry `json:"permissions"`
}
