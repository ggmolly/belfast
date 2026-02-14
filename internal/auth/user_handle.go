package auth

import (
	"crypto/rand"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
)

func EnsureUserHandle(user *orm.Account) error {
	if user == nil {
		return nil
	}
	if len(user.WebAuthnUserHandle) != 0 {
		return nil
	}
	handle := make([]byte, 32)
	if _, err := rand.Read(handle); err != nil {
		return err
	}
	user.WebAuthnUserHandle = handle
	return orm.UpdateAccountWebAuthnUserHandle(user.ID, handle, time.Now().UTC().Unix())
}
