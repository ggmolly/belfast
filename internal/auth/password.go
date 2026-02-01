package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"

	"github.com/ggmolly/belfast/internal/config"
)

const passwordAlgoArgon2id = "argon2id"

var (
	ErrPasswordTooShort = errors.New("password too short")
	ErrPasswordTooLong  = errors.New("password too long")
	ErrInvalidHash      = errors.New("invalid password hash")
)

func HashPassword(password string, cfg config.AuthConfig) (string, string, error) {
	if len(password) < cfg.PasswordMinLength {
		return "", "", ErrPasswordTooShort
	}
	if len(password) > cfg.PasswordMaxLength {
		return "", "", ErrPasswordTooLong
	}

	salt := make([]byte, cfg.PasswordHashParams.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", "", err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		cfg.PasswordHashParams.Iterations,
		cfg.PasswordHashParams.Memory,
		cfg.PasswordHashParams.Parallelism,
		cfg.PasswordHashParams.KeyLength,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	encoded := fmt.Sprintf("%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		passwordAlgoArgon2id,
		argon2.Version,
		cfg.PasswordHashParams.Memory,
		cfg.PasswordHashParams.Iterations,
		cfg.PasswordHashParams.Parallelism,
		b64Salt,
		b64Hash,
	)

	return encoded, passwordAlgoArgon2id, nil
}

func VerifyPassword(password string, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 5 {
		return false, ErrInvalidHash
	}
	if parts[0] != passwordAlgoArgon2id {
		return false, ErrInvalidHash
	}
	if !strings.HasPrefix(parts[1], "v=") {
		return false, ErrInvalidHash
	}
	var version int
	if _, err := fmt.Sscanf(parts[1], "v=%d", &version); err != nil {
		return false, ErrInvalidHash
	}
	if version != argon2.Version {
		return false, ErrInvalidHash
	}

	var memory uint32
	var iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[2], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false, ErrInvalidHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false, ErrInvalidHash
	}
	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, ErrInvalidHash
	}

	computed := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(decodedHash)))
	if subtle.ConstantTimeCompare(decodedHash, computed) != 1 {
		return false, nil
	}
	return true, nil
}
