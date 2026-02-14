package orm

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ggmolly/belfast/internal/authz"
	"github.com/ggmolly/belfast/internal/db"
)

var userRegistrationTestOnce sync.Once

func initUserRegistrationTest(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	userRegistrationTestOnce.Do(func() {
		InitDatabase()
	})
	ensurePlayerRole(t)
	clearTable(t, &AccountRole{})
	clearTable(t, &Account{})
	clearTable(t, &UserRegistrationChallenge{})
}

func ensurePlayerRole(t *testing.T) {
	t.Helper()
	now := time.Now().UTC()
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `
		INSERT INTO roles (id, name, description, created_at, updated_at, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (name) DO NOTHING
	`, uuid.NewString(), authz.RolePlayer, "player role", now, now, nil); err != nil {
		t.Fatalf("ensure player role failed: %v", err)
	}
}

func TestConsumeUserRegistrationChallengeByIDWithContextSuccess(t *testing.T) {
	initUserRegistrationTest(t)

	now := time.Now().UTC().Truncate(time.Second)
	challenge, err := CreateUserRegistrationChallenge(420001, "654321", "hash", "argon2", now.Add(time.Hour), now)
	if err != nil {
		t.Fatalf("create registration challenge failed: %v", err)
	}

	account, err := ConsumeUserRegistrationChallengeByIDWithContext(context.Background(), challenge.ID, "654321", now.Add(time.Minute))
	if err != nil {
		t.Fatalf("consume registration challenge failed: %v", err)
	}
	if account == nil || account.CommanderID == nil {
		t.Fatalf("expected created account with commander id")
	}
	if *account.CommanderID != 420001 {
		t.Fatalf("expected commander id 420001, got %d", *account.CommanderID)
	}

	stored, err := GetUserRegistrationChallenge(challenge.ID)
	if err != nil {
		t.Fatalf("get registration challenge failed: %v", err)
	}
	if stored.Status != UserRegistrationStatusConsumed {
		t.Fatalf("expected status %q, got %q", UserRegistrationStatusConsumed, stored.Status)
	}
	if stored.ConsumedAt == nil {
		t.Fatalf("expected consumed_at to be set")
	}
}

func TestConsumeUserRegistrationChallengeByIDWithContextPinMismatch(t *testing.T) {
	initUserRegistrationTest(t)

	now := time.Now().UTC().Truncate(time.Second)
	challenge, err := CreateUserRegistrationChallenge(420002, "123456", "hash", "argon2", now.Add(time.Hour), now)
	if err != nil {
		t.Fatalf("create registration challenge failed: %v", err)
	}

	_, err = ConsumeUserRegistrationChallengeByIDWithContext(context.Background(), challenge.ID, "000000", now.Add(time.Minute))
	if !errors.Is(err, ErrRegistrationChallengePinMismatch) {
		t.Fatalf("expected ErrRegistrationChallengePinMismatch, got %v", err)
	}

	stored, err := GetUserRegistrationChallenge(challenge.ID)
	if err != nil {
		t.Fatalf("get registration challenge failed: %v", err)
	}
	if stored.Status != UserRegistrationStatusPending {
		t.Fatalf("expected status %q, got %q", UserRegistrationStatusPending, stored.Status)
	}
}

func TestConsumeUserRegistrationChallengeByIDWithContextCancelled(t *testing.T) {
	initUserRegistrationTest(t)

	now := time.Now().UTC().Truncate(time.Second)
	challenge, err := CreateUserRegistrationChallenge(420003, "111111", "hash", "argon2", now.Add(time.Hour), now)
	if err != nil {
		t.Fatalf("create registration challenge failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = ConsumeUserRegistrationChallengeByIDWithContext(ctx, challenge.ID, "111111", now.Add(time.Minute))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestUpdateUserRegistrationChallengeStatusPreservesConsumedAt(t *testing.T) {
	initUserRegistrationTest(t)

	now := time.Now().UTC().Truncate(time.Second)
	challenge, err := CreateUserRegistrationChallenge(430001, "666999", "hash", "argon2", now.Add(time.Hour), now)
	if err != nil {
		t.Fatalf("create registration challenge failed: %v", err)
	}

	if _, err = ConsumeUserRegistrationChallengeByID(challenge.ID, "666999", now.Add(time.Minute)); err != nil {
		t.Fatalf("consume registration challenge failed: %v", err)
	}

	consumed, err := GetUserRegistrationChallenge(challenge.ID)
	if err != nil {
		t.Fatalf("load consumed challenge failed: %v", err)
	}
	if consumed.Status != UserRegistrationStatusConsumed || consumed.ConsumedAt == nil {
		t.Fatalf("expected consumed challenge with consumed_at, got status=%q consumed_at=nil?%v", consumed.Status, consumed.ConsumedAt == nil)
	}

	if err := UpdateUserRegistrationChallengeStatus(challenge.ID, UserRegistrationStatusExpired); err != nil {
		t.Fatalf("update status to expired failed: %v", err)
	}

	updated, err := GetUserRegistrationChallenge(challenge.ID)
	if err != nil {
		t.Fatalf("load updated challenge failed: %v", err)
	}
	if updated.Status != UserRegistrationStatusExpired {
		t.Fatalf("expected status %q, got %q", UserRegistrationStatusExpired, updated.Status)
	}
	if updated.ConsumedAt == nil || !updated.ConsumedAt.Equal(*consumed.ConsumedAt) {
		t.Fatalf("expected consumed_at to remain %v, got %v", consumed.ConsumedAt, updated.ConsumedAt)
	}
}

func TestUpdateUserRegistrationChallengeStatusSetsConsumedAtWhenConsumed(t *testing.T) {
	initUserRegistrationTest(t)

	now := time.Now().UTC().Truncate(time.Second)
	challenge, err := CreateUserRegistrationChallenge(430002, "111222", "hash", "argon2", now.Add(time.Hour), now)
	if err != nil {
		t.Fatalf("create registration challenge failed: %v", err)
	}

	if err := UpdateUserRegistrationChallengeStatus(challenge.ID, UserRegistrationStatusConsumed); err != nil {
		t.Fatalf("update status to consumed failed: %v", err)
	}

	stored, err := GetUserRegistrationChallenge(challenge.ID)
	if err != nil {
		t.Fatalf("load consumed challenge failed: %v", err)
	}
	if stored.Status != UserRegistrationStatusConsumed {
		t.Fatalf("expected status %q, got %q", UserRegistrationStatusConsumed, stored.Status)
	}
	if stored.ConsumedAt == nil {
		t.Fatalf("expected consumed_at to be set")
	}
}
