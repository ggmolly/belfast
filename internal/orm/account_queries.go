package orm

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ggmolly/belfast/internal/db"
)

func GetAccountByCommanderID(commanderID uint32) (*Account, error) {
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetAccountByCommanderID(ctx, pgInt8FromUint32Ptr(&commanderID))
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	account := Account{
		ID:                 row.ID,
		Username:           pgTextPtr(row.Username),
		UsernameNormalized: pgTextPtr(row.UsernameNormalized),
		CommanderID:        pgInt8PtrToUint32Ptr(row.CommanderID),
		PasswordHash:       row.PasswordHash,
		PasswordAlgo:       row.PasswordAlgo,
		PasswordUpdatedAt:  row.PasswordUpdatedAt.Time,
		IsAdmin:            row.IsAdmin,
		DisabledAt:         pgTimestamptzPtr(row.DisabledAt),
		LastLoginAt:        pgTimestamptzPtr(row.LastLoginAt),
		WebAuthnUserHandle: row.WebAuthnUserHandle,
		CreatedAt:          row.CreatedAt.Time,
		UpdatedAt:          row.UpdatedAt.Time,
	}
	return &account, nil
}

func UpdateAccountLastLoginAt(accountID string, loginAt time.Time) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE accounts
SET last_login_at = $2,
    updated_at = $2
WHERE id = $1
`, accountID, pgtype.Timestamptz{Time: loginAt, Valid: true})
	return err
}
