package answer

import (
	"time"

	"github.com/ggmolly/belfast/internal/orm"
	"golang.org/x/crypto/bcrypt"
)

const (
	secondaryPasswordLength         = 6
	secondaryPasswordMaxAttempts    = 5
	secondaryPasswordLockoutSeconds = 300
)

func isSecondaryPasswordValid(password string) bool {
	if len(password) != secondaryPasswordLength {
		return false
	}
	for i := 0; i < len(password); i++ {
		if password[i] < '0' || password[i] > '9' {
			return false
		}
	}
	return true
}

func hashSecondaryPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func verifySecondaryPassword(hash string, password string) bool {
	if hash == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func resetSecondaryPasswordLockoutIfExpired(commanderID uint32, settings orm.SecondaryPasswordSettings, now int64) (orm.SecondaryPasswordSettings, error) {
	if settings.FailCd == nil || now < *settings.FailCd {
		return settings, nil
	}
	if err := orm.ResetSecondaryPasswordLockout(orm.GormDB, commanderID); err != nil {
		return settings, err
	}
	settings.FailCd = nil
	settings.FailCount = 0
	return settings, nil
}

func recordSecondaryPasswordFailure(commanderID uint32, settings orm.SecondaryPasswordSettings, now int64) (orm.SecondaryPasswordSettings, error) {
	settings.FailCount++
	if settings.FailCount >= secondaryPasswordMaxAttempts {
		lockout := now + secondaryPasswordLockoutSeconds
		settings.FailCd = &lockout
	}
	if err := orm.UpdateSecondaryPasswordLockout(orm.GormDB, commanderID, settings.FailCount, settings.FailCd); err != nil {
		return settings, err
	}
	return settings, nil
}

func currentUnixTime() int64 {
	return time.Now().Unix()
}
