package orm

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/ggmolly/belfast/internal/db"
)

func ListExchangeCodes(offset int, limit int) ([]ExchangeCode, int64, error) {
	ctx := context.Background()
	offset, limit, unlimited := normalizePagination(offset, limit)

	var total int64
	if err := db.DefaultStore.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM exchange_codes`).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
SELECT id, code, platform, quota, rewards, created_at, updated_at
FROM exchange_codes
ORDER BY id ASC
OFFSET $1

`
	args := []any{int64(offset)}
	if !unlimited {
		query += `LIMIT $2`
		args = append(args, int64(limit))
	}
	rows, err := db.DefaultStore.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	codes := make([]ExchangeCode, 0)
	for rows.Next() {
		code, err := scanExchangeCode(rows)
		if err != nil {
			return nil, 0, err
		}
		codes = append(codes, code)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return codes, total, nil
}

func GetExchangeCode(codeID uint32) (*ExchangeCode, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT id, code, platform, quota, rewards, created_at, updated_at
FROM exchange_codes
WHERE id = $1
`, int64(codeID))
	code, err := scanExchangeCode(row)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &code, nil
}

func CreateExchangeCode(code *ExchangeCode) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO exchange_codes (code, platform, quota, rewards)
VALUES ($1, $2, $3, $4)
`, code.Code, code.Platform, code.Quota, code.Rewards)
	return err
}

func UpdateExchangeCode(code *ExchangeCode) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE exchange_codes
SET code = $2,
	platform = $3,
	quota = $4,
	rewards = $5,
	updated_at = CURRENT_TIMESTAMP
WHERE id = $1
`, int64(code.ID), code.Code, code.Platform, code.Quota, code.Rewards)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func DeleteExchangeCode(codeID uint32) error {
	ctx := context.Background()
	return WithPGXTx(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `DELETE FROM exchange_code_redeems WHERE exchange_code_id = $1`, int64(codeID)); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `DELETE FROM exchange_codes WHERE id = $1`, int64(codeID))
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return db.ErrNotFound
		}
		return nil
	})
}

func ListExchangeCodeRedeems(codeID uint32, offset int, limit int) ([]ExchangeCodeRedeem, int64, error) {
	ctx := context.Background()
	offset, limit, unlimited := normalizePagination(offset, limit)

	var total int64
	if err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM exchange_code_redeems
WHERE exchange_code_id = $1
`, int64(codeID)).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
SELECT exchange_code_id, commander_id, redeemed_at
FROM exchange_code_redeems
WHERE exchange_code_id = $1
ORDER BY redeemed_at DESC
OFFSET $2
`
	args := []any{int64(codeID), int64(offset)}
	if !unlimited {
		query += `LIMIT $3`
		args = append(args, int64(limit))
	}
	rows, err := db.DefaultStore.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	redeems := make([]ExchangeCodeRedeem, 0)
	for rows.Next() {
		var (
			redeem         ExchangeCodeRedeem
			exchangeCodeID int64
			commanderID    int64
		)
		if err := rows.Scan(&exchangeCodeID, &commanderID, &redeem.RedeemedAt); err != nil {
			return nil, 0, err
		}
		redeem.ExchangeCodeID = uint32(exchangeCodeID)
		redeem.CommanderID = uint32(commanderID)
		redeems = append(redeems, redeem)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return redeems, total, nil
}

func CreateExchangeCodeRedeem(codeID uint32, commanderID uint32, redeemedAt time.Time) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO exchange_code_redeems (exchange_code_id, commander_id, redeemed_at)
VALUES ($1, $2, $3)
`, int64(codeID), int64(commanderID), redeemedAt)
	return err
}

func DeleteExchangeCodeRedeem(codeID uint32, commanderID uint32) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM exchange_code_redeems
WHERE exchange_code_id = $1
  AND commander_id = $2
`, int64(codeID), int64(commanderID))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "23505"
}

func scanExchangeCode(scanner rowScanner) (ExchangeCode, error) {
	var (
		code ExchangeCode
		id   int64
	)
	if err := scanner.Scan(&id, &code.Code, &code.Platform, &code.Quota, &code.Rewards, &code.CreatedAt, &code.UpdatedAt); err != nil {
		return ExchangeCode{}, err
	}
	code.ID = uint32(id)
	return code, nil
}
