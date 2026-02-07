package types

type UserAccount struct {
	ID          string `json:"id"`
	CommanderID uint32 `json:"commander_id"`
	Disabled    bool   `json:"disabled"`
	LastLoginAt string `json:"last_login_at"`
	CreatedAt   string `json:"created_at"`
}

type UserSession struct {
	ID        string `json:"id"`
	ExpiresAt string `json:"expires_at"`
}

type UserAuthLoginRequest struct {
	CommanderID uint32 `json:"commander_id" validate:"required,gt=0"`
	Password    string `json:"password" validate:"required"`
}

type UserAuthLoginResponse struct {
	User    UserAccount `json:"user"`
	Session UserSession `json:"session"`
}

type UserAuthSessionResponse struct {
	User      UserAccount `json:"user"`
	Session   UserSession `json:"session"`
	CSRFToken string      `json:"csrf_token"`
}

type UserRegistrationChallengeRequest struct {
	CommanderID uint32 `json:"commander_id" validate:"required,gt=0"`
	Password    string `json:"password" validate:"required"`
}

type UserRegistrationChallengeResponse struct {
	ChallengeID string `json:"challenge_id"`
	ExpiresAt   string `json:"expires_at"`
}

type UserRegistrationStatusResponse struct {
	Status string `json:"status"`
}

type UserRegistrationVerifyRequest struct {
	Pin string `json:"pin" validate:"required"`
}
