package orm

import (
	"context"

	"github.com/ggmolly/belfast/internal/db"
)

func GetLocalAccountByAccount(account string) (*LocalAccount, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT arg2, account, password, mail_box, created_at, updated_at
FROM local_accounts
WHERE account = $1
`, account)
	var entry LocalAccount
	err := row.Scan(&entry.Arg2, &entry.Account, &entry.Password, &entry.MailBox, &entry.CreatedAt, &entry.UpdatedAt)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func UpdateLocalAccountPassword(arg2 uint32, password string) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE local_accounts
SET password = $2,
    updated_at = now()
WHERE arg2 = $1
`, int64(arg2), password)
	return err
}

func CreateLocalAccount(entry LocalAccount) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO local_accounts (
  arg2,
  account,
  password,
  mail_box,
  created_at,
  updated_at
) VALUES (
  $1, $2, $3, $4, now(), now()
)
`, int64(entry.Arg2), entry.Account, entry.Password, entry.MailBox)
	return err
}

func LocalArg2Exists(value uint32) (bool, error) {
	ctx := context.Background()
	var exists bool
	if err := db.DefaultStore.Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM local_accounts WHERE arg2 = $1)`, int64(value)).Scan(&exists); err != nil {
		return false, err
	}
	if exists {
		return true, nil
	}
	if err := db.DefaultStore.Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM yostarus_maps WHERE arg2 = $1)`, int64(value)).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}
