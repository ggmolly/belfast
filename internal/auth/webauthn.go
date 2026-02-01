package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/ggmolly/belfast/internal/config"
)

type WebAuthnProvider interface {
	BeginRegistration(user webauthn.User, opts ...webauthn.RegistrationOption) (*protocol.CredentialCreation, *webauthn.SessionData, error)
	FinishRegistration(user webauthn.User, session webauthn.SessionData, r *http.Request) (*webauthn.Credential, error)
	BeginLogin(user webauthn.User, opts ...webauthn.LoginOption) (*protocol.CredentialAssertion, *webauthn.SessionData, error)
	BeginDiscoverableLogin(opts ...webauthn.LoginOption) (*protocol.CredentialAssertion, *webauthn.SessionData, error)
	FinishLogin(user webauthn.User, session webauthn.SessionData, r *http.Request) (*webauthn.Credential, error)
	FinishPasskeyLogin(handler webauthn.DiscoverableUserHandler, session webauthn.SessionData, r *http.Request) (webauthn.User, *webauthn.Credential, error)
}

type Manager struct {
	Config    config.AuthConfig
	WebAuthn  WebAuthnProvider
	Limiter   *RateLimiter
	Selection protocol.AuthenticatorSelection
}

func NewManager(cfg config.AuthConfig) (*Manager, error) {
	cfg = NormalizeConfig(cfg)
	selection := protocol.AuthenticatorSelection{UserVerification: protocol.VerificationPreferred}
	manager := &Manager{
		Config:    cfg,
		Limiter:   NewRateLimiter(),
		Selection: selection,
	}
	if cfg.WebAuthnRPID == "" || cfg.WebAuthnRPName == "" || len(cfg.WebAuthnExpectedOrigins) == 0 {
		return manager, nil
	}

	webConfig := &webauthn.Config{
		RPID:                   cfg.WebAuthnRPID,
		RPDisplayName:          cfg.WebAuthnRPName,
		RPOrigins:              cfg.WebAuthnExpectedOrigins,
		AuthenticatorSelection: selection,
		Timeouts: webauthn.TimeoutsConfig{
			Registration: webauthn.TimeoutConfig{
				Enforce:    true,
				Timeout:    time.Duration(cfg.WebAuthnChallengeTTLSeconds) * time.Second,
				TimeoutUVD: time.Duration(cfg.WebAuthnChallengeTTLSeconds) * time.Second,
			},
			Login: webauthn.TimeoutConfig{
				Enforce:    true,
				Timeout:    time.Duration(cfg.WebAuthnChallengeTTLSeconds) * time.Second,
				TimeoutUVD: time.Duration(cfg.WebAuthnChallengeTTLSeconds) * time.Second,
			},
		},
	}
	web, err := webauthn.New(webConfig)
	if err != nil {
		return nil, err
	}
	manager.WebAuthn = &realWebAuthn{web: web}
	return manager, nil
}

func (manager *Manager) EnsureWebAuthn() error {
	if manager == nil || manager.WebAuthn == nil {
		return errors.New("webauthn is not configured")
	}
	return nil
}

type realWebAuthn struct {
	web *webauthn.WebAuthn
}

func (r *realWebAuthn) BeginRegistration(user webauthn.User, opts ...webauthn.RegistrationOption) (*protocol.CredentialCreation, *webauthn.SessionData, error) {
	return r.web.BeginRegistration(user, opts...)
}

func (r *realWebAuthn) FinishRegistration(user webauthn.User, session webauthn.SessionData, req *http.Request) (*webauthn.Credential, error) {
	return r.web.FinishRegistration(user, session, req)
}

func (r *realWebAuthn) BeginLogin(user webauthn.User, opts ...webauthn.LoginOption) (*protocol.CredentialAssertion, *webauthn.SessionData, error) {
	return r.web.BeginLogin(user, opts...)
}

func (r *realWebAuthn) BeginDiscoverableLogin(opts ...webauthn.LoginOption) (*protocol.CredentialAssertion, *webauthn.SessionData, error) {
	return r.web.BeginDiscoverableLogin(opts...)
}

func (r *realWebAuthn) FinishLogin(user webauthn.User, session webauthn.SessionData, req *http.Request) (*webauthn.Credential, error) {
	return r.web.FinishLogin(user, session, req)
}

func (r *realWebAuthn) FinishPasskeyLogin(handler webauthn.DiscoverableUserHandler, session webauthn.SessionData, req *http.Request) (webauthn.User, *webauthn.Credential, error) {
	return r.web.FinishPasskeyLogin(handler, session, req)
}
