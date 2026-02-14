package orm

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ggmolly/belfast/internal/authz"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

const (
	UserRegistrationStatusPending  = "pending"
	UserRegistrationStatusConsumed = "consumed"
	UserRegistrationStatusExpired  = "expired"
)

var (
	ErrUserAccountExists                = errors.New("user account exists")
	ErrRegistrationChallengeExists      = errors.New("registration challenge exists")
	ErrRegistrationChallengeNotFound    = errors.New("registration challenge not found")
	ErrRegistrationChallengeConsumed    = errors.New("registration challenge consumed")
	ErrRegistrationChallengeExpired     = errors.New("registration challenge expired")
	ErrRegistrationChallengeMismatch    = errors.New("registration challenge mismatch")
	ErrRegistrationChallengePinMismatch = errors.New("registration challenge pin mismatch")
	ErrRegistrationPinExists            = errors.New("registration pin exists")
)

func CreateUserRegistrationChallenge(commanderID uint32, pin string, passwordHash string, passwordAlgo string, expiresAt time.Time, now time.Time) (*UserRegistrationChallenge, error) {
	ctx := context.Background()
	existingCount, err := db.DefaultStore.Queries.CountAccountsByCommanderID(ctx, pgInt8FromUint32Ptr(&commanderID))
	if err != nil {
		return nil, err
	}
	if existingCount > 0 {
		return nil, ErrUserAccountExists
	}

	pending, err := db.DefaultStore.Queries.GetLatestPendingRegistrationChallengeByCommander(ctx, gen.GetLatestPendingRegistrationChallengeByCommanderParams{CommanderID: int64(commanderID), Status: UserRegistrationStatusPending})
	err = db.MapNotFound(err)
	if err == nil {
		_ = db.DefaultStore.Queries.UpdateRegistrationChallengeStatus(ctx, gen.UpdateRegistrationChallengeStatusParams{ID: pending.ID, Status: UserRegistrationStatusExpired, ConsumedAt: pending.ConsumedAt})
	} else if !db.IsNotFound(err) {
		return nil, err
	}

	_, err = db.DefaultStore.Queries.GetPendingRegistrationChallengeByPin(ctx, gen.GetPendingRegistrationChallengeByPinParams{Pin: pin, Status: UserRegistrationStatusPending, ExpiresAt: pgtype.Timestamptz{Time: now, Valid: true}})
	err = db.MapNotFound(err)
	if err == nil {
		return nil, ErrRegistrationPinExists
	} else if !db.IsNotFound(err) {
		return nil, err
	}

	entry := UserRegistrationChallenge{
		ID:           uuid.NewString(),
		CommanderID:  commanderID,
		Pin:          pin,
		PasswordHash: passwordHash,
		PasswordAlgo: passwordAlgo,
		Status:       UserRegistrationStatusPending,
		ExpiresAt:    expiresAt,
		CreatedAt:    now,
	}
	if err := db.DefaultStore.Queries.CreateRegistrationChallenge(ctx, gen.CreateRegistrationChallengeParams{
		ID:           entry.ID,
		CommanderID:  int64(entry.CommanderID),
		Pin:          entry.Pin,
		PasswordHash: entry.PasswordHash,
		PasswordAlgo: entry.PasswordAlgo,
		Status:       entry.Status,
		ExpiresAt:    pgtype.Timestamptz{Time: entry.ExpiresAt, Valid: true},
		ConsumedAt:   pgtype.Timestamptz{},
		CreatedAt:    pgtype.Timestamptz{Time: entry.CreatedAt, Valid: true},
	}); err != nil {
		return nil, err
	}
	return &entry, nil
}

func ConsumeUserRegistrationChallengeWithContext(ctx context.Context, commanderID uint32, pin string, now time.Time) (*Account, error) {
	var account *Account
	err := db.DefaultStore.WithTx(ctx, func(q *gen.Queries) error {
		role, err := q.GetRoleByName(ctx, authz.RolePlayer)
		if err != nil {
			return err
		}
		challenge, err := q.GetPendingRegistrationChallengeByPinForUpdate(ctx, gen.GetPendingRegistrationChallengeByPinForUpdateParams{Pin: pin, Status: UserRegistrationStatusPending})
		err = db.MapNotFound(err)
		if err != nil {
			if db.IsNotFound(err) {
				return ErrRegistrationChallengeNotFound
			}
			return err
		}
		if uint32(challenge.CommanderID) != commanderID {
			return ErrRegistrationChallengeMismatch
		}
		if !challenge.ExpiresAt.Time.After(now) {
			_ = q.UpdateRegistrationChallengeStatus(ctx, gen.UpdateRegistrationChallengeStatusParams{ID: challenge.ID, Status: UserRegistrationStatusExpired, ConsumedAt: challenge.ConsumedAt})
			return ErrRegistrationChallengeExpired
		}

		_, err = q.GetAccountByCommanderID(ctx, pgInt8FromUint32Ptr(&commanderID))
		err = db.MapNotFound(err)
		if err == nil {
			return ErrUserAccountExists
		} else if !db.IsNotFound(err) {
			return err
		}

		created := Account{
			ID:                uuid.NewString(),
			CommanderID:       &commanderID,
			PasswordHash:      challenge.PasswordHash,
			PasswordAlgo:      challenge.PasswordAlgo,
			PasswordUpdatedAt: now,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		if err := q.CreateAccount(ctx, gen.CreateAccountParams{ID: created.ID, CommanderID: pgInt8FromUint32Ptr(created.CommanderID), PasswordHash: created.PasswordHash, PasswordAlgo: created.PasswordAlgo, PasswordUpdatedAt: pgtype.Timestamptz{Time: created.PasswordUpdatedAt, Valid: true}, CreatedAt: pgtype.Timestamptz{Time: created.CreatedAt, Valid: true}, UpdatedAt: pgtype.Timestamptz{Time: created.UpdatedAt, Valid: true}}); err != nil {
			return err
		}
		if err := q.CreateAccountRoleLink(ctx, gen.CreateAccountRoleLinkParams{AccountID: created.ID, RoleID: role.ID, CreatedAt: pgtype.Timestamptz{Time: now, Valid: true}}); err != nil {
			return err
		}
		if err := q.UpdateRegistrationChallengeStatus(ctx, gen.UpdateRegistrationChallengeStatusParams{ID: challenge.ID, Status: UserRegistrationStatusConsumed, ConsumedAt: pgtype.Timestamptz{Time: now, Valid: true}}); err != nil {
			return err
		}
		account = &created
		return nil
	})
	if err != nil {
		return nil, err
	}
	return account, nil
}

func ConsumeUserRegistrationChallenge(commanderID uint32, pin string, now time.Time) (*Account, error) {
	return ConsumeUserRegistrationChallengeWithContext(context.Background(), commanderID, pin, now)
}

func ConsumeUserRegistrationChallengeByIDWithContext(ctx context.Context, id string, pin string, now time.Time) (*Account, error) {
	var account *Account
	err := db.DefaultStore.WithTx(ctx, func(q *gen.Queries) error {
		role, err := q.GetRoleByName(ctx, authz.RolePlayer)
		if err != nil {
			return err
		}
		challenge, err := q.GetRegistrationChallengeByIDForUpdate(ctx, id)
		err = db.MapNotFound(err)
		if err != nil {
			if db.IsNotFound(err) {
				return ErrRegistrationChallengeNotFound
			}
			return err
		}
		switch challenge.Status {
		case UserRegistrationStatusConsumed:
			return ErrRegistrationChallengeConsumed
		case UserRegistrationStatusExpired:
			return ErrRegistrationChallengeExpired
		}
		if !challenge.ExpiresAt.Time.After(now) {
			_ = q.UpdateRegistrationChallengeStatus(ctx, gen.UpdateRegistrationChallengeStatusParams{ID: challenge.ID, Status: UserRegistrationStatusExpired, ConsumedAt: challenge.ConsumedAt})
			return ErrRegistrationChallengeExpired
		}
		if challenge.Pin != pin {
			return ErrRegistrationChallengePinMismatch
		}

		commanderIDArg := pgtype.Int8{Int64: challenge.CommanderID, Valid: true}
		_, err = q.GetAccountByCommanderID(ctx, commanderIDArg)
		err = db.MapNotFound(err)
		if err == nil {
			return ErrUserAccountExists
		} else if !db.IsNotFound(err) {
			return err
		}

		commanderID := uint32(challenge.CommanderID)
		created := Account{
			ID:                uuid.NewString(),
			CommanderID:       &commanderID,
			PasswordHash:      challenge.PasswordHash,
			PasswordAlgo:      challenge.PasswordAlgo,
			PasswordUpdatedAt: now,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		if err := q.CreateAccount(ctx, gen.CreateAccountParams{ID: created.ID, CommanderID: pgInt8FromUint32Ptr(created.CommanderID), PasswordHash: created.PasswordHash, PasswordAlgo: created.PasswordAlgo, PasswordUpdatedAt: pgtype.Timestamptz{Time: created.PasswordUpdatedAt, Valid: true}, CreatedAt: pgtype.Timestamptz{Time: created.CreatedAt, Valid: true}, UpdatedAt: pgtype.Timestamptz{Time: created.UpdatedAt, Valid: true}}); err != nil {
			return err
		}
		if err := q.CreateAccountRoleLink(ctx, gen.CreateAccountRoleLinkParams{AccountID: created.ID, RoleID: role.ID, CreatedAt: pgtype.Timestamptz{Time: now, Valid: true}}); err != nil {
			return err
		}
		if err := q.UpdateRegistrationChallengeStatus(ctx, gen.UpdateRegistrationChallengeStatusParams{ID: challenge.ID, Status: UserRegistrationStatusConsumed, ConsumedAt: pgtype.Timestamptz{Time: now, Valid: true}}); err != nil {
			return err
		}
		account = &created
		return nil
	})
	if err != nil {
		return nil, err
	}
	return account, nil
}

func ConsumeUserRegistrationChallengeByID(id string, pin string, now time.Time) (*Account, error) {
	return ConsumeUserRegistrationChallengeByIDWithContext(context.Background(), id, pin, now)
}

func GetUserRegistrationChallenge(id string) (*UserRegistrationChallenge, error) {
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetRegistrationChallengeByID(ctx, id)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	challenge := UserRegistrationChallenge{
		ID:           row.ID,
		CommanderID:  uint32(row.CommanderID),
		Pin:          row.Pin,
		PasswordHash: row.PasswordHash,
		PasswordAlgo: row.PasswordAlgo,
		Status:       row.Status,
		ExpiresAt:    row.ExpiresAt.Time,
		ConsumedAt:   pgTimestamptzPtr(row.ConsumedAt),
		CreatedAt:    row.CreatedAt.Time,
	}
	return &challenge, nil
}

func UpdateUserRegistrationChallengeStatus(id string, status string) error {
	ctx := context.Background()
	challenge, err := db.DefaultStore.Queries.GetRegistrationChallengeByID(ctx, id)
	err = db.MapNotFound(err)
	if db.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	consumedAt := challenge.ConsumedAt
	if status == UserRegistrationStatusConsumed && !consumedAt.Valid {
		consumedAt = pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	}
	return db.DefaultStore.Queries.UpdateRegistrationChallengeStatus(ctx, gen.UpdateRegistrationChallengeStatusParams{ID: id, Status: status, ConsumedAt: consumedAt})
}
