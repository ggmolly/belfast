package auth

import (
	"testing"

	"github.com/ggmolly/belfast/internal/config"
)

func TestHashAndVerifyPassword(t *testing.T) {
	cfg := NormalizeConfig(config.AuthConfig{})
	hash, algo, err := HashPassword("this-is-a-strong-password", cfg)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if algo != passwordAlgoArgon2id {
		t.Fatalf("expected algo %s, got %s", passwordAlgoArgon2id, algo)
	}
	ok, err := VerifyPassword("this-is-a-strong-password", hash)
	if err != nil {
		t.Fatalf("verify password: %v", err)
	}
	if !ok {
		t.Fatalf("expected password to verify")
	}
	ok, err = VerifyPassword("wrong-password", hash)
	if err != nil {
		t.Fatalf("verify wrong password: %v", err)
	}
	if ok {
		t.Fatalf("expected wrong password to fail")
	}
}

func TestHashPasswordTooShort(t *testing.T) {
	cfg := NormalizeConfig(config.AuthConfig{})
	_, _, err := HashPassword("short", cfg)
	if err == nil {
		t.Fatalf("expected error for short password")
	}
	if err != ErrPasswordTooShort {
		t.Fatalf("expected ErrPasswordTooShort, got %v", err)
	}
}
