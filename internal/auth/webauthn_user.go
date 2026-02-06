package auth

import (
	"encoding/base64"
	"errors"
	"strconv"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/ggmolly/belfast/internal/orm"
)

type WebAuthnUser struct {
	ID          []byte
	Name        string
	DisplayName string
	Credentials []webauthn.Credential
}

func (user WebAuthnUser) WebAuthnID() []byte {
	return user.ID
}

func (user WebAuthnUser) WebAuthnName() string {
	return user.Name
}

func (user WebAuthnUser) WebAuthnDisplayName() string {
	return user.DisplayName
}

func (user WebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	return user.Credentials
}

func BuildWebAuthnUser(account orm.Account, credentials []orm.WebAuthnCredential) (WebAuthnUser, error) {
	if len(account.WebAuthnUserHandle) == 0 {
		return WebAuthnUser{}, errors.New("missing webauthn user handle")
	}
	webauthnCredentials := make([]webauthn.Credential, 0, len(credentials))
	for _, credential := range credentials {
		built, err := buildWebAuthnCredential(credential)
		if err != nil {
			return WebAuthnUser{}, err
		}
		webauthnCredentials = append(webauthnCredentials, built)
	}
	name := ""
	if account.Username != nil {
		name = *account.Username
	}
	if name == "" && account.CommanderID != nil {
		name = "commander:" + strconv.FormatUint(uint64(*account.CommanderID), 10)
	}
	if name == "" {
		name = account.ID
	}
	return WebAuthnUser{
		ID:          account.WebAuthnUserHandle,
		Name:        name,
		DisplayName: name,
		Credentials: webauthnCredentials,
	}, nil
}

func buildWebAuthnCredential(record orm.WebAuthnCredential) (webauthn.Credential, error) {
	id, err := base64.RawURLEncoding.DecodeString(record.CredentialID)
	if err != nil {
		return webauthn.Credential{}, err
	}
	var aaguid []byte
	if record.AAGUID != "" {
		if aaguid, err = base64.RawURLEncoding.DecodeString(record.AAGUID); err != nil {
			return webauthn.Credential{}, err
		}
	}
	transports := make([]protocol.AuthenticatorTransport, 0, len(record.Transports))
	for _, transport := range record.Transports {
		transports = append(transports, protocol.AuthenticatorTransport(transport))
	}
	flags := webauthn.CredentialFlags{}
	if record.BackupEligible != nil {
		flags.BackupEligible = *record.BackupEligible
	}
	if record.BackupState != nil {
		flags.BackupState = *record.BackupState
	}
	return webauthn.Credential{
		ID:              id,
		PublicKey:       record.PublicKey,
		AttestationType: record.AttestationFmt,
		Transport:       transports,
		Flags:           flags,
		Authenticator: webauthn.Authenticator{
			AAGUID:    aaguid,
			SignCount: record.SignCount,
		},
	}, nil
}
