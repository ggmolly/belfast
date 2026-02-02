package answer

import (
	"sort"

	"github.com/ggmolly/belfast/internal/auth"
	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/orm"
)

const (
	secondaryPasswordLength         = 6
	secondaryPasswordMaxFailures    = 5
	secondaryPasswordLockoutSeconds = 300
)

func secondaryPasswordConfig() config.AuthConfig {
	cfg := config.AuthConfig{
		PasswordMinLength: secondaryPasswordLength,
		PasswordMaxLength: secondaryPasswordLength,
	}
	return auth.NormalizeConfig(cfg)
}

func hashSecondaryPassword(password string) (string, error) {
	hash, _, err := auth.HashPassword(password, secondaryPasswordConfig())
	if err != nil {
		return "", err
	}
	return hash, nil
}

func verifySecondaryPassword(password string, encoded string) (bool, error) {
	return auth.VerifyPassword(password, encoded)
}

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

func sanitizeSecondarySystemList(values []uint32) []uint32 {
	if len(values) == 0 {
		return []uint32{}
	}
	unique := make(map[uint32]struct{}, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		unique[value] = struct{}{}
	}
	if len(unique) == 0 {
		return []uint32{}
	}
	list := make([]uint32, 0, len(unique))
	for value := range unique {
		list = append(list, value)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i] < list[j]
	})
	return list
}

func secondaryPasswordLocked(state *orm.SecondaryPasswordState, now uint32) bool {
	return state.FailCd > 0 && now < state.FailCd
}

func applySecondaryPasswordFailure(state *orm.SecondaryPasswordState, now uint32) {
	if state.FailCount < secondaryPasswordMaxFailures {
		state.FailCount++
	}
	if state.FailCount >= secondaryPasswordMaxFailures {
		state.FailCount = secondaryPasswordMaxFailures
		state.FailCd = now + secondaryPasswordLockoutSeconds
	}
}
