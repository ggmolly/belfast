package orm

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ggmolly/belfast/internal/db"
)

func CountAdminAccounts() (int64, error) {
	ctx := context.Background()
	var count int64
	err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM accounts
WHERE is_admin = true
   OR EXISTS (
	SELECT 1
	FROM account_roles
	JOIN roles ON roles.id = account_roles.role_id
	WHERE account_roles.account_id = accounts.id
	  AND roles.name = 'admin'
)
`).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func ListAdminAccounts(offset int, limit int) ([]Account, int64, error) {
	ctx := context.Background()

	var total int64
	if err := db.DefaultStore.Pool.QueryRow(ctx, `
	SELECT COUNT(*)
	FROM accounts
	WHERE is_admin = true
	   OR EXISTS (
		SELECT 1
		FROM account_roles
		JOIN roles ON roles.id = account_roles.role_id
		WHERE account_roles.account_id = accounts.id
		  AND roles.name = 'admin'
	)
	`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := db.DefaultStore.Pool.Query(ctx, `
	SELECT id,
	       username,
       username_normalized,
       commander_id,
       password_hash,
       password_algo,
       password_updated_at,
       is_admin,
       disabled_at,
       last_login_at,
       web_authn_user_handle,
       created_at,
       updated_at
FROM accounts
	WHERE is_admin = true
	   OR EXISTS (
		SELECT 1
		FROM account_roles
		JOIN roles ON roles.id = account_roles.role_id
		WHERE account_roles.account_id = accounts.id
		  AND roles.name = 'admin'
	)
	ORDER BY created_at DESC
	OFFSET $1
	LIMIT $2
`, int64(offset), int64(limit))
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	accounts := make([]Account, 0)
	for rows.Next() {
		account, err := scanAccountRow(rows)
		if err != nil {
			return nil, 0, err
		}
		accounts = append(accounts, account)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return accounts, total, nil
}

func GetAccountByID(accountID string) (*Account, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT id,
       username,
       username_normalized,
       commander_id,
       password_hash,
       password_algo,
       password_updated_at,
       is_admin,
       disabled_at,
       last_login_at,
       web_authn_user_handle,
       created_at,
       updated_at
FROM accounts
WHERE id = $1
`, accountID)
	account, err := scanAccountRow(row)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func GetAccountByUsernameNormalized(usernameNormalized string) (*Account, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT id,
       username,
       username_normalized,
       commander_id,
       password_hash,
       password_algo,
       password_updated_at,
       is_admin,
       disabled_at,
       last_login_at,
       web_authn_user_handle,
       created_at,
       updated_at
FROM accounts
WHERE username_normalized = $1
`, usernameNormalized)
	account, err := scanAccountRow(row)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func GetAccountByWebAuthnUserHandle(handle []byte) (*Account, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT id,
       username,
       username_normalized,
       commander_id,
       password_hash,
       password_algo,
       password_updated_at,
       is_admin,
       disabled_at,
       last_login_at,
       web_authn_user_handle,
       created_at,
       updated_at
FROM accounts
WHERE web_authn_user_handle = $1
`, handle)
	account, err := scanAccountRow(row)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func CreateAccount(account *Account) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO accounts (
	id,
	username,
	username_normalized,
	commander_id,
	password_hash,
	password_algo,
	password_updated_at,
	is_admin,
	disabled_at,
	last_login_at,
	web_authn_user_handle,
	created_at,
	updated_at
)
VALUES (
	$1,
	$2,
	$3,
	$4,
	$5,
	$6,
	$7,
	$8,
	$9,
	$10,
	$11,
	$12,
	$13
)
`,
		account.ID,
		pgTextFromPtr(account.Username),
		pgTextFromPtr(account.UsernameNormalized),
		pgInt8FromUint32Ptr(account.CommanderID),
		account.PasswordHash,
		account.PasswordAlgo,
		pgTimestamptz(account.PasswordUpdatedAt),
		account.IsAdmin,
		pgTimestamptzFromPtr(account.DisabledAt),
		pgTimestamptzFromPtr(account.LastLoginAt),
		account.WebAuthnUserHandle,
		pgTimestamptz(account.CreatedAt),
		pgTimestamptz(account.UpdatedAt),
	)
	return err
}

func UpdateAccountUsername(accountID string, username string, usernameNormalized string, updatedAt time.Time) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE accounts
SET username = $2,
	username_normalized = $3,
	updated_at = $4
WHERE id = $1
`, accountID, username, usernameNormalized, pgTimestamptz(updatedAt))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func UpdateAccountDisabledAt(accountID string, disabledAt *time.Time, updatedAt time.Time) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE accounts
SET disabled_at = $2,
	updated_at = $3
WHERE id = $1
`, accountID, pgTimestamptzFromPtr(disabledAt), pgTimestamptz(updatedAt))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func UpdateAccountPassword(accountID string, passwordHash string, passwordAlgo string, passwordUpdatedAt time.Time, updatedAt time.Time) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE accounts
SET password_hash = $2,
	password_algo = $3,
	password_updated_at = $4,
	updated_at = $5
WHERE id = $1
`, accountID, passwordHash, passwordAlgo, pgTimestamptz(passwordUpdatedAt), pgTimestamptz(updatedAt))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func DeleteAccountByID(accountID string) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM accounts WHERE id = $1`, accountID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func AccountUsernameNormalizedExists(usernameNormalized string, excludeAccountID string) (bool, error) {
	ctx := context.Background()
	var count int64
	if excludeAccountID == "" {
		err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM accounts
WHERE username_normalized = $1
`, usernameNormalized).Scan(&count)
		if err != nil {
			return false, err
		}
	} else {
		err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM accounts
WHERE username_normalized = $1
  AND id <> $2
`, usernameNormalized, excludeAccountID).Scan(&count)
		if err != nil {
			return false, err
		}
	}
	return count > 0, nil
}

func CountEnabledAccountsWithRole(roleName string, excludeAccountID string) (int64, error) {
	ctx := context.Background()
	var count int64
	if excludeAccountID == "" {
		err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM account_roles
JOIN roles ON roles.id = account_roles.role_id
JOIN accounts ON accounts.id = account_roles.account_id
WHERE roles.name = $1
  AND accounts.disabled_at IS NULL
`, roleName).Scan(&count)
		if err != nil {
			return 0, err
		}
	} else {
		err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM account_roles
JOIN roles ON roles.id = account_roles.role_id
JOIN accounts ON accounts.id = account_roles.account_id
WHERE roles.name = $1
  AND accounts.disabled_at IS NULL
  AND accounts.id <> $2
`, roleName, excludeAccountID).Scan(&count)
		if err != nil {
			return 0, err
		}
	}
	return count, nil
}

func ListWebAuthnCredentialsByUserID(userID string) ([]WebAuthnCredential, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT id,
       user_id,
       credential_id,
       public_key,
       sign_count,
       transports,
       aaguid,
       attestation_fmt,
       resident_key,
       backup_eligible,
       backup_state,
       created_at,
       last_used_at,
       label,
       rp_id
FROM web_authn_credentials
WHERE user_id = $1
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	credentials := make([]WebAuthnCredential, 0)
	for rows.Next() {
		record, err := scanWebAuthnCredentialRow(rows)
		if err != nil {
			return nil, err
		}
		credentials = append(credentials, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return credentials, nil
}

func CreateWebAuthnCredential(record *WebAuthnCredential) error {
	ctx := context.Background()
	transports, err := json.Marshal(record.Transports)
	if err != nil {
		return err
	}
	_, err = db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO web_authn_credentials (
	id,
	user_id,
	credential_id,
	public_key,
	sign_count,
	transports,
	aaguid,
	attestation_fmt,
	resident_key,
	backup_eligible,
	backup_state,
	created_at,
	last_used_at,
	label,
	rp_id
)
VALUES (
	$1,
	$2,
	$3,
	$4,
	$5,
	$6,
	$7,
	$8,
	$9,
	$10,
	$11,
	$12,
	$13,
	$14,
	$15
)
`,
		record.ID,
		record.UserID,
		record.CredentialID,
		record.PublicKey,
		int64(record.SignCount),
		transports,
		record.AAGUID,
		record.AttestationFmt,
		residentKeyToText(record.ResidentKey),
		pgBoolFromPtr(record.BackupEligible),
		pgBoolFromPtr(record.BackupState),
		pgTimestamptz(record.CreatedAt),
		pgTimestamptzFromPtr(record.LastUsedAt),
		pgTextFromPtr(record.Label),
		record.RPID,
	)
	return err
}

func DeleteWebAuthnCredentialByUserAndCredentialID(userID string, credentialID string) (bool, error) {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM web_authn_credentials
WHERE user_id = $1
  AND credential_id = $2
`, userID, credentialID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func DeleteWebAuthnCredentialsByUserID(userID string) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM web_authn_credentials WHERE user_id = $1`, userID)
	return err
}

func WebAuthnCredentialExists(credentialID string) (bool, error) {
	ctx := context.Background()
	var exists bool
	err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT EXISTS(
	SELECT 1
	FROM web_authn_credentials
	WHERE credential_id = $1
)
`, credentialID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func WebAuthnCredentialExistsForUser(userID string, credentialID string) (bool, error) {
	ctx := context.Background()
	var exists bool
	err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT EXISTS(
	SELECT 1
	FROM web_authn_credentials
	WHERE user_id = $1
	  AND credential_id = $2
)
`, userID, credentialID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func UpdateWebAuthnCredentialUsageByCredentialID(credentialID string, signCount uint32, lastUsedAt time.Time, backupEligible *bool, backupState *bool) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE web_authn_credentials
SET sign_count = $2,
	last_used_at = $3,
	backup_eligible = $4,
	backup_state = $5
WHERE credential_id = $1
`, credentialID, int64(signCount), pgTimestamptz(lastUsedAt), pgBoolFromPtr(backupEligible), pgBoolFromPtr(backupState))
	return err
}

func scanAccountRow(scanner rowScanner) (Account, error) {
	var (
		account            Account
		username           pgtype.Text
		usernameNormalized pgtype.Text
		commanderID        pgtype.Int8
		disabledAt         pgtype.Timestamptz
		lastLoginAt        pgtype.Timestamptz
		webAuthnHandle     []byte
	)
	if err := scanner.Scan(
		&account.ID,
		&username,
		&usernameNormalized,
		&commanderID,
		&account.PasswordHash,
		&account.PasswordAlgo,
		&account.PasswordUpdatedAt,
		&account.IsAdmin,
		&disabledAt,
		&lastLoginAt,
		&webAuthnHandle,
		&account.CreatedAt,
		&account.UpdatedAt,
	); err != nil {
		return Account{}, err
	}
	account.Username = pgTextPtr(username)
	account.UsernameNormalized = pgTextPtr(usernameNormalized)
	commanderIDValue, convErr := pgInt8PtrToUint32PtrChecked(commanderID)
	if convErr != nil {
		return Account{}, convErr
	}
	account.CommanderID = commanderIDValue
	account.DisabledAt = pgTimestamptzPtr(disabledAt)
	account.LastLoginAt = pgTimestamptzPtr(lastLoginAt)
	account.WebAuthnUserHandle = append([]byte(nil), webAuthnHandle...)
	return account, nil
}

func scanWebAuthnCredentialRow(scanner rowScanner) (WebAuthnCredential, error) {
	var (
		record         WebAuthnCredential
		signCount      int64
		transportsJSON []byte
		backupEligible pgtype.Bool
		backupState    pgtype.Bool
		lastUsedAt     pgtype.Timestamptz
		label          pgtype.Text
		residentKey    pgtype.Text
	)
	if err := scanner.Scan(
		&record.ID,
		&record.UserID,
		&record.CredentialID,
		&record.PublicKey,
		&signCount,
		&transportsJSON,
		&record.AAGUID,
		&record.AttestationFmt,
		&residentKey,
		&backupEligible,
		&backupState,
		&record.CreatedAt,
		&lastUsedAt,
		&label,
		&record.RPID,
	); err != nil {
		return WebAuthnCredential{}, err
	}
	convertedSignCount, convErr := Uint32FromInt64Checked(signCount)
	if convErr != nil {
		return WebAuthnCredential{}, convErr
	}
	record.SignCount = convertedSignCount
	record.BackupEligible = pgBoolPtr(backupEligible)
	record.BackupState = pgBoolPtr(backupState)
	record.LastUsedAt = pgTimestamptzPtr(lastUsedAt)
	record.Label = pgTextPtr(label)
	record.ResidentKey = residentKeyFromText(residentKey)
	if len(transportsJSON) != 0 {
		if err := json.Unmarshal(transportsJSON, &record.Transports); err != nil {
			return WebAuthnCredential{}, err
		}
	}
	return record, nil
}

func residentKeyToText(value bool) string {
	if value {
		return "required"
	}
	return ""
}

func residentKeyFromText(value pgtype.Text) bool {
	if !value.Valid {
		return false
	}
	return strings.TrimSpace(value.String) != ""
}
