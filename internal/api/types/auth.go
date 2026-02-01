package types

type AdminUser struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	IsAdmin     bool   `json:"is_admin"`
	Disabled    bool   `json:"disabled"`
	LastLoginAt string `json:"last_login_at"`
	CreatedAt   string `json:"created_at"`
}

type AuthSession struct {
	ID        string `json:"id"`
	ExpiresAt string `json:"expires_at"`
}

type AuthSessionResponse struct {
	User      AdminUser   `json:"user"`
	Session   AuthSession `json:"session"`
	CSRFToken string      `json:"csrf_token"`
}

type AuthLoginResponse struct {
	User    AdminUser   `json:"user"`
	Session AuthSession `json:"session"`
}

type AuthBootstrapRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthPasswordChangeRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type AdminUserListResponse struct {
	Users []AdminUser    `json:"users"`
	Meta  PaginationMeta `json:"meta"`
}

type AdminUserResponse struct {
	User AdminUser `json:"user"`
}

type AdminUserCreateRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AdminUserUpdateRequest struct {
	Username *string `json:"username,omitempty"`
	Disabled *bool   `json:"disabled,omitempty"`
}

type AdminUserPasswordUpdateRequest struct {
	Password string `json:"password"`
}

type PasskeyRegisterOptionsRequest struct {
	Label            *string `json:"label,omitempty"`
	UserVerification *string `json:"user_verification,omitempty"`
	ResidentKey      *string `json:"resident_key,omitempty"`
}

type PasskeyAuthenticateOptionsRequest struct {
	Username *string `json:"username,omitempty"`
}

type PasskeyAttestationResponse struct {
	ClientDataJSON    string `json:"clientDataJSON"`
	AttestationObject string `json:"attestationObject"`
}

type PasskeyRegistrationCredential struct {
	ID       string                     `json:"id"`
	RawID    string                     `json:"rawId"`
	Type     string                     `json:"type"`
	Response PasskeyAttestationResponse `json:"response"`
}

type PasskeyRegisterVerifyRequest struct {
	Credential PasskeyRegistrationCredential `json:"credential"`
	Label      *string                       `json:"label,omitempty"`
}

type PasskeyAssertionResponse struct {
	ClientDataJSON    string `json:"clientDataJSON"`
	AuthenticatorData string `json:"authenticatorData"`
	Signature         string `json:"signature"`
	UserHandle        string `json:"userHandle,omitempty"`
}

type PasskeyAuthenticationCredential struct {
	ID       string                   `json:"id"`
	RawID    string                   `json:"rawId"`
	Type     string                   `json:"type"`
	Response PasskeyAssertionResponse `json:"response"`
}

type PasskeyAuthenticateVerifyRequest struct {
	Credential PasskeyAuthenticationCredential `json:"credential"`
	Username   *string                         `json:"username,omitempty"`
}

type PasskeySummary struct {
	CredentialID   string   `json:"credential_id"`
	Label          string   `json:"label"`
	CreatedAt      string   `json:"created_at"`
	LastUsedAt     string   `json:"last_used_at"`
	Transports     []string `json:"transports"`
	AAGUID         string   `json:"aaguid"`
	BackupEligible *bool    `json:"backup_eligible"`
	BackupState    *bool    `json:"backup_state"`
}

type PasskeyListResponse struct {
	Passkeys []PasskeySummary `json:"passkeys"`
}

type PasskeyRegisterResponse struct {
	CredentialID string `json:"credential_id"`
	Label        string `json:"label"`
	CreatedAt    string `json:"created_at"`
}
